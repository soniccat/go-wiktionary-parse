package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/macdub/go-colorlog"
)

const (
	WikitextElementTypeText     = 1
	WikitextElementTypeMarkup   = 2
	WikitextElementTypeSection  = 3
	WikitextElementTypeTemplate = 4
	WikitextElementTypeNewline  = 5
)

var WiktionaryErrorLogger *colorlog.ColorLog

type Wikitext struct {
	elements []WikitextElement
}

func (ws *Wikitext) addElement(e WikitextElement) {
	ws.elements = append(ws.elements, e)
}

type WikitextElement interface {
	ElementType() int
}

type WikitextSectionElement struct {
	level int
	name  string
}

func (e *WikitextSectionElement) ElementType() int {
	return WikitextElementTypeSection
}

func parseSectionElement(reader *strings.Reader) (WikitextSectionElement, error) {
	// parse ==English==
	element := WikitextSectionElement{}
	readingFirstPart := true
	readingText := false
	nameBuilder := strings.Builder{}

	var r rune
	var err error

	for {
		r, _, err = reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			} else {
				break
			}

		} else if r == rune('=') {
			if readingFirstPart {
				element.level += 1
			} else {
				if readingText {
					readingText = false
				} // else - just skip the rune
			}
		} else if r == rune('\n') {
			// skip and don't create WikitextNewlineElement here
			break
		} else {
			if readingFirstPart || readingText {
				if readingFirstPart {
					readingFirstPart = false
					readingText = true
				}

				nameBuilder.WriteRune(r)
			} else {
				reader.UnreadRune()
				break
			}
		}
	}

	element.name = strings.Trim(nameBuilder.String(), " ")

	return element, err
}

type WikitextTemplateElement struct {
	name  string
	props []WikitextTemplateProp
}

func (wt *WikitextTemplateElement) addProp(p WikitextTemplateProp) {
	wt.props = append(wt.props, p)
}

func (e *WikitextTemplateElement) ElementType() int {
	return WikitextElementTypeTemplate
}

func (e *WikitextTemplateElement) PropByName(name string) *WikitextTemplateProp {
	for _, v := range e.props {
		if v.name == name {
			return &v
		}
	}

	return nil
}

// to get these |1= |2= |3=, anI starts with 0 which means "|1="
func (e *WikitextTemplateElement) PropStringPropByIndex(anI int) *WikitextTemplateProp {
	i := 0
	for _, v := range e.props {
		if v.isStringValue() {
			if i == anI {
				return &v
			}

			i += 1
		}
	}

	return nil
}

type WikitextTemplateProp struct {
	name  string
	value *string // TODO: store list of elements (WikitextTemplateElement, WikitextTextElement)
}

func (e *WikitextTemplateProp) isStringValue() bool {
	return e.value == nil
}

func (e *WikitextTemplateProp) stringValue() string {
	return e.name
}

func (e *WikitextTemplateProp) isInnerStringValue() bool {
	return !e.isStringValue() && e.value != nil
}

func (e *WikitextTemplateProp) innerStringValue() string {
	return *e.value
}

type WikitextNewlineElement struct {
}

func (e *WikitextNewlineElement) ElementType() int {
	return WikitextElementTypeNewline
}

func parseTemplateElement(reader *strings.Reader) (WikitextTemplateElement, error) {
	// parse {{quote-text|en|year=2002|author=w:John Fusco|title={{w|Spirit: Stallion of the Cimarron}}|passage=Colonel: See, gentlemen? Any horse could be '''broken'''.}}
	element := WikitextTemplateElement{}
	nameBuilder := strings.Builder{}

	var r rune
	var err error
	readingFirstPart := true
	readingName := false
	readingProp := false
	isInvalidState := false

	for {
		r, _, err = reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			} else {
				break
			}

		} else if r == rune('{') { // expect {{
			if readingFirstPart {
				r, _, err = reader.ReadRune()
				if r == rune('{') {
					readingFirstPart = false
					readingName = true
				} else {
					isInvalidState = true
					break
				}
			} else if readingName {
				bstr, _ := peek(reader, 1)
				if strings.HasPrefix(bstr, "{") {
					reader.UnreadByte()

					// important: inner templates would be converted to text
					innerElement := WikitextTemplateElement{}
					innerElement, err = parseTemplateElement(reader)
					if err != nil {
						break
					}

					var addStr string
					if len(innerElement.props) == 0 {
						addStr = innerElement.name
					} else {
						p0 := innerElement.props[0]
						if p0.isStringValue() {
							addStr = p0.stringValue()
						} else if p0.isInnerStringValue() {
							addStr = p0.innerStringValue()
						}
					}
					nameBuilder.WriteString(addStr)

				} else {
					// read sth like {2}
					nameBuilder.WriteRune(r)
					for { // TODO: move that in a separate function
						var r2 rune
						r2, _, err = reader.ReadRune()
						if err != nil {
							break
						}
						nameBuilder.WriteRune(r2)
						if r2 == '}' {
							break
						}
					}

					if err != nil {
						break
					}
				}
			} else {
				isInvalidState = true
				break
			}

		} else if r == rune('|') {
			if readingName || readingProp {
				readingName = false
				readingProp = true
				var prop WikitextTemplateProp
				prop, err = parseTemplateProp(reader)
				if err != nil {
					isInvalidState = true
					break
				} else {
					element.addProp(prop)
				}
			} else {
				isInvalidState = true
				break
			}

		} else if r == '}' { // expect }}
			r, _, err = reader.ReadRune()
			if r == '}' {
				break
			} else {
				isInvalidState = true
				break
			}
		} else {
			if readingName {
				nameBuilder.WriteRune(r)
			} else if r == '\n' {
				// skip and don't create WikitextNewlineElement here
			} else {
				isInvalidState = true
				break
			}
		}
	}

	element.name = nameBuilder.String()

	if err != nil || isInvalidState {
		return element, fmt.Errorf("parseTemplateElement: unexpected '%v' (%w)", string(r), err)
	}

	return element, nil
}

func parseTemplateProp(reader *strings.Reader) (WikitextTemplateProp, error) {
	// parse title={{w|Spirit: Stallion of the Cimarron}}
	// important: name link will be converted to link name ([[Rail (magazine)|Rail]] -> Rail), so wiktionary link text will be lost here
	element := WikitextTemplateProp{}
	nameBuilder := strings.Builder{}
	valueStringBuilder := strings.Builder{}

	var r rune
	var err error
	readingName := true
	readingStringValue := false
	isInvalidState := false

	var writeF func(str string) = func(str string) {
		if readingName {
			nameBuilder.WriteString(str)
		} else if readingStringValue {
			valueStringBuilder.WriteString(str)
		}
	}

	for {
		r, _, err = reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			} else {
				break
			}

		} else if r == rune('=') {
			if readingName {
				readingName = false
				readingStringValue = true
			} else {
				valueStringBuilder.WriteRune(r)
			}

		} else if r == '|' || r == '}' {
			isSeparatorProcessed := false
			if r == '|' {
				bstr, _ := peek(reader, 4)
				if strings.HasPrefix(bstr, "_|") {
					writeF(" ")
					_, err = reader.Seek(2, io.SeekCurrent)
					if err == nil {
						isSeparatorProcessed = true
					}
				} else if strings.HasPrefix(bstr, "or|") {
					writeF(" or ")
					_, err = reader.Seek(3, io.SeekCurrent)
					if err == nil {
						isSeparatorProcessed = true
					}
				} else if strings.HasPrefix(bstr, "and|") {
					writeF(" and ")
					_, err = reader.Seek(4, io.SeekCurrent)
					if err == nil {
						isSeparatorProcessed = true
					}
				} else if strings.HasPrefix(bstr, ";|") {
					writeF("; ")
					_, err = reader.Seek(2, io.SeekCurrent)
					if err == nil {
						isSeparatorProcessed = true
					}
				}
			}

			bstr, _ := peek(reader, 1)
			if err != nil {
				isInvalidState = true
				break
			} else if !isSeparatorProcessed && r == '}' {
				if strings.HasPrefix(bstr, "}") {
					reader.UnreadByte()
					break
				} else {
					writeF(string(r))
				}
			} else if !isSeparatorProcessed {
				reader.UnreadByte()
				break
			}
		} else {
			var isProcessed = false
			var addStr string

			addStr, isProcessed, err = parseWikitextTextBlock(r, reader, "")
			if err != nil {
				break
			}

			bstr, _ := peek(reader, 1)
			if !isProcessed && r == '{' && strings.HasPrefix(bstr, "{") {
				reader.UnreadByte()

				// important: inner templates would be converted to text
				innerElement := WikitextTemplateElement{}
				innerElement, err = parseTemplateElement(reader)
				if err != nil {
					break
				}

				if len(innerElement.props) == 0 {
					addStr = innerElement.name
				} else {
					p0 := innerElement.props[0]
					if p0.isStringValue() {
						addStr = p0.stringValue()
					} else if p0.isInnerStringValue() {
						addStr = p0.innerStringValue()
					}
				}
				isProcessed = true
			}

			if isProcessed {
				if len(addStr) > 0 {
					writeF(addStr)
				}
			} else {
				writeF(string(r))
			}
		}
	}

	element.name = nameBuilder.String()
	if readingStringValue {
		element.value = Ptr(valueStringBuilder.String())
	}

	if err != nil || isInvalidState {
		return element, fmt.Errorf("parseTemplateProp: unexpected '%v' (%w)", string(r), err)
	}

	return element, nil
}

type WikitextLink struct {
	text string
	name string
}

func parseWikitextLink(reader *strings.Reader) (link WikitextLink, err error) {
	linkTextBuilder := strings.Builder{}
	linkNameBuilder := strings.Builder{}

	var r rune
	var c int
	var isReadingLinkText = true
	var isReadingLinkName = false
	var gotToEnd = false
	var isInvalidState bool
	var readByteCount int

	r, c, err = reader.ReadRune()
	readByteCount += c
	if err != nil {
		return
	}

	if r != '[' {
		err = errors.New("parseWikitextLink: first rune isn't '['")
		return
	}

	for {
		r, c, err = reader.ReadRune()
		readByteCount += c
		if err != nil {
			break
		}

		if r == ']' {
			nextR, _ := peek(reader, 1)
			if nextR == "]" {
				reader.Seek(1, io.SeekCurrent)
				gotToEnd = true
				break
			} else {
				if isReadingLinkText {
					linkTextBuilder.WriteRune(r)
				} else if isReadingLinkName {
					linkNameBuilder.WriteRune(r)
				} else {
					break
				}
			}
		} else if r == '|' && isReadingLinkText {
			isReadingLinkText = false
			isReadingLinkName = true
		} else if r == '\n' {
			break // handle wrong formatting
		} else {
			if isReadingLinkText {
				linkTextBuilder.WriteRune(r)
			} else if isReadingLinkName {
				linkNameBuilder.WriteRune(r)
			}
		}
	}

	link.text = linkTextBuilder.String()
	link.name = linkNameBuilder.String()

	if len(link.name) == 0 {
		link.name = link.text
	}

	if err != nil || isInvalidState || !gotToEnd {
		if err == io.EOF || !gotToEnd {
			// not a link, seek to the initial position
			reader.Seek(int64(-readByteCount), io.SeekCurrent)
			return link, notWikitextLink{}
		}

		return link, fmt.Errorf("parseWikitextLink: unexpected '%v' (%w)", string(r), err)
	}

	return
}

type notWikitextLink struct {
}

func (n notWikitextLink) Error() string {
	return "not wikitextlink"
}

type WikitextTextElement struct {
	value string
}

func (e *WikitextTextElement) ElementType() int {
	return WikitextElementTypeText
}

type WikitextMarkupElement struct {
	value string
}

func (e *WikitextMarkupElement) ElementType() int {
	return WikitextElementTypeMarkup
}

func parseWikitextTextBlock(r rune, reader *strings.Reader, exclude string) (s string, isHandled bool, err error) {
	bstr, _ := peek(reader, 7)
	if r == '[' && strings.HasPrefix(bstr, "[") {
		if exclude == "[[" {
			return
		}
		// important: the link text will be lost, here only the link name is kept
		var l WikitextLink
		l, err = parseWikitextLink(reader)
		if errors.Is(err, notWikitextLink{}) {
			err = nil
			return
		}

		s = l.name
		isHandled = true
		if err != nil {
			return
		}
		// } else if len(exclude) > 0 && exclude == string(r)+bstr[0:len(exclude)-1] {
		// 	return
	} else if r == '<' && strings.HasPrefix(bstr, "math>") {
		if exclude == "<math>" {
			return
		}
		reader.Seek(5, io.SeekCurrent)
		s, err = readUntil(reader, "</math>")
		isHandled = true

	} else if r == '<' && strings.HasPrefix(bstr, "br>") {
		if exclude == "<br>" {
			return
		}
		reader.Seek(3, io.SeekCurrent)
		s = "\n"
		isHandled = true

		/*} else if r == '\'' && strings.HasPrefix(bstr, "''''") {
		if exclude == "'''''" {
			return
		}
		reader.Seek(4, io.SeekCurrent)
		s, err = readUntil(reader, "'''''")
		isHandled = true
		*/
	} else if r == '\'' && strings.HasPrefix(bstr, "''") {
		if exclude == "'''" {
			return
		}
		reader.Seek(2, io.SeekCurrent)
		s = ""
		//s, err = readUntil(reader, "'''")
		isHandled = true

	} else if r == '\'' && strings.HasPrefix(bstr, "'") {
		if exclude == "''" {
			return
		}
		reader.Seek(1, io.SeekCurrent)
		s = ""
		// s, err = readUntil(reader, "''")
		isHandled = true
	}

	return
}

func parseWikitext(str string) (Wikitext, error) {
	wikitext := Wikitext{}
	reader := strings.NewReader(str)
	markupBuilder := strings.Builder{}
	textBuilder := strings.Builder{}

	isNewLine := true
	hasMarkupInLine := false
	readingMarkup := false
	readingText := false

	var r rune
	var err error
	var parsedElement1 WikitextElement
	var parsedElement2 WikitextElement

	for {
		r = rune(0)
		isProcessed := false
		bstr, _ := peek(reader, 2)
		canRead := len(bstr) > 0

		// handle special cases
		if isNewLine && strings.HasPrefix(bstr, "=") {
			el, e := parseSectionElement(reader)
			parsedElement2 = Ptr(el)
			err = e
			isProcessed = true
		} else if strings.HasPrefix(bstr, "{{") {
			el, e := parseTemplateElement(reader)
			parsedElement2 = Ptr(el)
			err = e
			isProcessed = true
		}

		// read next character
		isMarkup := false
		isNewLine = false
		if !isProcessed && canRead {
			r, _, err = reader.ReadRune()
			if !hasMarkupInLine {
				isMarkup = r == '*' || r == '#' || r == ':'
			}
			isNewLine = r == '\n'
			if isNewLine {
				hasMarkupInLine = false
			}
		}

		// save read text in an element if needed
		if readingText && (isProcessed || isMarkup || !canRead || isNewLine) {
			textElement := WikitextTextElement{}
			textElement.value = strings.Trim(textBuilder.String(), " ")
			textBuilder.Reset()
			parsedElement1 = Ptr(textElement)
			readingText = false
		}

		if readingMarkup && (isProcessed || !isMarkup || !canRead || isNewLine) {
			markupElement := WikitextMarkupElement{}
			markupElement.value = strings.Trim(markupBuilder.String(), " ")
			markupBuilder.Reset()
			parsedElement1 = Ptr(markupElement)
			readingMarkup = false
			hasMarkupInLine = !isNewLine
		}

		// append read character in a buffer
		if !isProcessed && isMarkup && err == nil && canRead {
			readingMarkup = true
			markupBuilder.WriteRune(r)
		}

		if !isProcessed && !isMarkup && err == nil && canRead && !isNewLine {
			if r == ' ' && !readingText {
				// skip
			} else {
				var addStr string
				var isHandled bool
				addStr, isHandled, err = parseWikitextTextBlock(r, reader, "")
				if isHandled {
					readingText = true
					textBuilder.WriteString(addStr)
				}
				if err != nil {
					break
				}
				if !isHandled {
					readingText = true
					textBuilder.WriteRune(r)
				}
			}
		}

		// handle results, add created elements in the right order
		if err != nil {
			l, _ := readLine(reader)
			err = fmt.Errorf("error while parsing at \""+string(l)+"\": %w", err)
			break
		} else {
			if parsedElement1 != nil {
				wikitext.addElement(parsedElement1)
				parsedElement1 = nil
			}
			if parsedElement2 != nil {
				wikitext.addElement(parsedElement2)
				parsedElement2 = nil
			}
			if isNewLine {
				wikitext.addElement(&WikitextNewlineElement{})
			}
		}

		if !canRead {
			break
		}
	}

	return wikitext, err
}

// Tools

func Ptr[T any](x T) *T {
	return &x
}

func peek(reader *strings.Reader, l int) (string, error) {
	b := make([]byte, l)
	c, err := reader.Read(b)
	if err != nil {
		return "", err
	}

	bstr := string(b)
	_, err = reader.Seek(int64(-c), io.SeekCurrent)
	if err != nil {
		return bstr, err
	}

	return bstr, nil
}

func readLine(reader *strings.Reader) (string, error) {
	b := strings.Builder{}
	for {
		bt, err := reader.ReadByte()
		if err != nil {
			break
		} else if bt == '\n' {
			break
		} else {
			b.WriteByte(bt)
		}
	}

	return b.String(), nil
}

func readUntil(reader *strings.Reader, stop string) (string, error) {
	b := strings.Builder{}
	var err error
	for {
		var bt rune
		bt, _, err = reader.ReadRune()
		if err != nil {
			break

		} else if bt == '\n' {
			break // handle wrong formatting

		} else {
			var parsedStr string
			var isHandled bool
			parsedStr, isHandled, err = parseWikitextTextBlock(bt, reader, stop)
			if isHandled {
				b.WriteString(parsedStr)
			} else {

				if strings.HasPrefix(stop, string(bt)) {
					if len(stop) == 1 {
						b.WriteRune(bt)
						break
					}

					var lastPart string
					lastPart, err = peek(reader, len(stop)-len(string(bt)))
					if lastPart == stop[len(string(bt)):] {
						reader.Seek(int64(len(lastPart)), io.SeekCurrent)
						break
					}

					b.WriteString(lastPart)
					if err != nil {
						break
					}
				} else {
					b.WriteRune(bt)
				}
			}

			if err != nil {
				break
			}
		}
	}

	return b.String(), err
}
