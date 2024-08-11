package main

import (
	"bufio"
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
)

var WiktionaryErrorLogger *colorlog.ColorLog

type Wikitext struct {
	strings []WikitextString
}

func (wt *Wikitext) addString(w WikitextString) {
	wt.strings = append(wt.strings, w)
}

type WikitextString struct {
	elements []WikitextElement
}

func (ws *WikitextString) addElement(e WikitextElement) {
	ws.elements = append(ws.elements, e)
}

type WikitextElement interface {
	elementType() int
}

type WikitextSectionElement struct {
	level int
	name  string
}

func (e *WikitextSectionElement) elementType() int {
	return WikitextElementTypeSection
}

func parseSectionElement(reader *bufio.Reader) (WikitextSectionElement, error) {
	// parse ==English==
	element := WikitextSectionElement{}
	readingFirstPart := true
	readingText := false
	nameBuilder := strings.Builder{}

	for {
		r, _, err := reader.ReadRune()
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

	return element, nil
}

type WikiTemplateElement struct {
	name  string
	props []WikiTemplateProp
}

func (wt *WikiTemplateElement) addProp(p WikiTemplateProp) {
	wt.props = append(wt.props, p)
}

type WikiTemplateProp struct {
	name  string
	value WikiTemplateElement
}

func (e *WikiTemplateElement) elementType() int {
	return WikitextElementTypeTemplate
}

func parseTemplateElement(reader *bufio.Reader) (WikiTemplateElement, error) {
	// parse {{quote-text|en|year=2002|author=w:John Fusco|title={{w|Spirit: Stallion of the Cimarron}}|passage=Colonel: See, gentlemen? Any horse could be '''broken'''.}}
	element := WikiTemplateElement{}
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
				prop, err := parseTemplateProp(reader)
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

func parseTemplateProp(reader *bufio.Reader) (WikiTemplateProp, error) {
	// parse title={{w|Spirit: Stallion of the Cimarron}}
	element := WikiTemplateProp{}
	nameBuilder := strings.Builder{}
	valueTemplateElement := WikiTemplateElement{}
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

				b, _ := reader.Peek(2)
				if len(b) > 1 && strings.HasPrefix(string(b), "{{") {
					valueTemplateElement, err = parseTemplateElement(reader)
					break
				} else {
					readingStringValue = true
				}
			} else {
				isInvalidState = true
				break
			}

		} else if r == '|' || r == '}' {
			reader.UnreadRune()
			break

		} else {
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

	element.name = nameBuilder.String()

	if readingStringValue {
		valueTemplateElement.name = valueStringBuilder.String()
	}
	element.value = valueTemplateElement

	if err != nil || isInvalidState {
		return element, fmt.Errorf("parseTemplateProp: unexpected '%v' (%w)", string(r), err)
	}

	return element, nil
}

func parseWikitext(text string) (Wikitext, error) {
	wikiText := Wikitext{}
	reader := bufio.NewReader(strings.NewReader(text))

	var str string
	var err error

	for {
		// read line by line
		str, err = reader.ReadString('\n')
		if len(str) > 0 {
			wikiString, err := parseWikiTextString(str)
			if err == nil {
				wikiText.addString(wikiString)
			}
		}

		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
	}

	return wikiText, err
}

type WikiTextElement struct {
	value string
}

func (e *WikiTextElement) elementType() int {
	return WikitextElementTypeText
}

type WikiMarkupElement struct {
	value string
}

func (e *WikiMarkupElement) elementType() int {
	return WikitextElementTypeMarkup
}

func parseWikiTextString(str string) (WikitextString, error) {
	wikiTextString := WikitextString{}
	reader := bufio.NewReader(strings.NewReader(str))
	markupBuilder := strings.Builder{}
	textBuilder := strings.Builder{}
	isFirst := true
	readingMarkup := false
	readingText := false

	var r rune
	var err error
	var parsedElement1 WikitextElement
	var parsedElement2 WikitextElement

	for {
		isProcessed := false
		b, _ := reader.Peek(2)
		if len(b) > 0 {
			substr := string(b)

			if isFirst && strings.HasPrefix(substr, "=") {
				el, e := parseSectionElement(reader)
				parsedElement2 = Ptr(el)
				err = e
				isProcessed = true
			} else if strings.HasPrefix(substr, "{{") {
				el, e := parseTemplateElement(reader)
				parsedElement2 = Ptr(el)
				err = e
				isProcessed = true
			}

			isMarkup := false
			if !isProcessed {
				r, _, err = reader.ReadRune()
				isMarkup = r == '*' || r == '#' || r == ':'
			}

			if isProcessed && readingText || isMarkup && readingText {
				textElement := WikiTextElement{}
				textElement.value = strings.Trim(textBuilder.String(), " ")
				textBuilder.Reset()
				parsedElement1 = Ptr(textElement)
				readingText = false
			}

			if isProcessed && readingMarkup || !isMarkup && readingMarkup {
				markupElement := WikiMarkupElement{}
				markupElement.value = strings.Trim(markupBuilder.String(), " ")
				markupBuilder.Reset()
				parsedElement1 = Ptr(markupElement)
				readingMarkup = false
			}

			if !isProcessed && isMarkup && err == nil {
				readingMarkup = true
				markupBuilder.WriteRune(r)
			}

			if !isProcessed && !isMarkup && err == nil {
				if r == '\n' {
					break
				} else if r == ' ' {
					continue
				}
				readingText = true
				textBuilder.WriteRune(r)
			}

			if err != nil {
				err = fmt.Errorf("error while parsing \""+str+"\": %w", err.Error())
				break
			} else {
				if parsedElement1 != nil {
					wikiTextString.addElement(parsedElement1)
					parsedElement1 = nil
				}
				if parsedElement2 != nil {
					wikiTextString.addElement(parsedElement2)
					parsedElement2 = nil
				}
			}

			isFirst = false
		} else {
			break
		}
	}

	return wikiTextString, err
}

// Tools

func Ptr[T any](x T) *T {
	return &x
}
