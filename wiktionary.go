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
	value *WikitextTemplateElement
}

func (e *WikitextTemplateProp) isStringValue() bool {
	return e.value == nil
}

func (e *WikitextTemplateProp) stringValue() string {
	return e.name
}

func (e *WikitextTemplateProp) isInnerStringValue() bool {
	return !e.isStringValue() && len(e.value.props) == 0
}

func (e *WikitextTemplateProp) innerStringValue() string {
	return e.value.name
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
	valueTemplateElement := WikitextTemplateElement{}
	valueStringBuilder := strings.Builder{}

	var r rune
	var err error
	readingName := true
	readingStringValue := false
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

		} else if r == rune('=') {
			if readingName {
				readingName = false

				str, _ := Peek(reader, 7)
				if strings.HasPrefix(str, "{{...}}") {
					valueStringBuilder.WriteString("...")
					reader.Seek(7, io.SeekCurrent)
					readingStringValue = true
				} else if strings.HasPrefix(str, "{{") {
					valueTemplateElement, err = parseTemplateElement(reader)
					break
				} else {
					readingStringValue = true
				}
			} else {
				if readingStringValue {
					valueStringBuilder.WriteRune(r)
				} else {
					isInvalidState = true
					break
				}
			}

		} else if r == '|' || r == '}' {
			isSeparatorProcessed := false
			if r == '|' {
				bstr, _ := Peek(reader, 4)
				if strings.HasPrefix(bstr, "_|") {
					nameBuilder.WriteString(" ")
					_, err = reader.Seek(2, io.SeekCurrent)
					if err == nil {
						isSeparatorProcessed = true
					}
				} else if strings.HasPrefix(bstr, "or|") {
					nameBuilder.WriteString(" or ")
					_, err = reader.Seek(3, io.SeekCurrent)
					if err == nil {
						isSeparatorProcessed = true
					}
				} else if strings.HasPrefix(bstr, "and|") {
					nameBuilder.WriteString(" and ")
					_, err = reader.Seek(4, io.SeekCurrent)
					if err == nil {
						isSeparatorProcessed = true
					}
				} else if strings.HasPrefix(bstr, ";|") {
					nameBuilder.WriteString("; ")
					_, err = reader.Seek(2, io.SeekCurrent)
					if err == nil {
						isSeparatorProcessed = true
					}
				}
			}

			if err != nil {
				isInvalidState = true
				break
			} else if !isSeparatorProcessed {
				reader.UnreadByte()
				break
			}
		} else {
			var isProcessed = false

			bstr, _ := Peek(reader, 6)
			if readingName && r == '[' && strings.HasPrefix(bstr, "[") {
				l, err := parseWikitextLink(reader)
				if err != nil {
					break
				}

				nameBuilder.WriteString(l.name)
				isProcessed = true
			} else if readingStringValue && r == '{' && strings.HasPrefix(bstr, "{...}}") {
				valueStringBuilder.WriteString("...")
				reader.Seek(6, io.SeekCurrent)
				isProcessed = true
			}

			if !isProcessed {
				if readingName {
					nameBuilder.WriteRune(r)
				} else if readingStringValue {
					valueStringBuilder.WriteRune(r)
				} else {
					isInvalidState = true
					break
				}
			}
		}
	}

	element.name = nameBuilder.String()

	if readingStringValue {
		valueTemplateElement.name = valueStringBuilder.String()
	}

	if len(valueTemplateElement.name) > 0 {
		element.value = &valueTemplateElement
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
	var isReadingLinkText = true
	var isReadingLinkName = false
	var isInvalidState bool

	r, _, err = reader.ReadRune()
	if err != nil {
		return
	}

	if r != '[' {
		err = errors.New("parseWikitextLink: first rune isn't '['")
		return
	}

	for {
		r, _, err = reader.ReadRune()
		if err != nil {
			break
		}

		if r == ']' {
			nextR, _ := Peek(reader, 1)
			if nextR == "]" {
				reader.Seek(1, io.SeekCurrent)
				break
			} else {
				isInvalidState = true
			}
		} else if r == '|' && isReadingLinkText {
			isReadingLinkText = false
			isReadingLinkName = true
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

	if err != nil || isInvalidState {
		return link, fmt.Errorf("parseWikitextLink: unexpected '%v' (%w)", string(r), err)
	}

	return
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

func parseWikitext(str string) (Wikitext, error) {
	wikitext := Wikitext{}
	reader := strings.NewReader(str)
	markupBuilder := strings.Builder{}
	textBuilder := strings.Builder{}

	isNewLine := true
	readingMarkup := false
	readingText := false

	var r rune
	var err error
	var parsedElement1 WikitextElement
	var parsedElement2 WikitextElement

	for {
		r = rune(0)
		isProcessed := false
		bstr, _ := Peek(reader, 2)
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
			isMarkup = r == '*' || r == '#' || r == ':'
			isNewLine = r == '\n'
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
				readingText = true
				textBuilder.WriteRune(r)
			}
		}

		// handle results, add created elements in the right order
		if err != nil {
			l, _ := ReadLine(reader)
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

func Peek(reader *strings.Reader, l int) (string, error) {
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

func ReadLine(reader *strings.Reader) (string, error) {
	b := strings.Builder{}
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			break
		} else {
			b.WriteRune(r)
		}
	}

	return b.String(), nil
}
