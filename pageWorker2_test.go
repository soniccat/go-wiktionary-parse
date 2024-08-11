package main

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsingSection1(t *testing.T) {
	str := "=t="
	e, err := parseSectionElement(bufio.NewReader(strings.NewReader(str)))

	assert.Nil(t, err)
	assert.Equal(t, "t", e.name)
	assert.Equal(t, 1, e.level)
}

func TestParsingSection2(t *testing.T) {
	str := "==ttt=="
	e, err := parseSectionElement(bufio.NewReader(strings.NewReader(str)))

	assert.Nil(t, err)
	assert.Equal(t, "ttt", e.name)
	assert.Equal(t, 2, e.level)
}

func TestParsingTemplateProp1(t *testing.T) {
	str := "a=d"
	e, err := parseTemplateProp(bufio.NewReader(strings.NewReader(str)))

	assert.Nil(t, err)
	assert.Equal(t, "a", e.name)
	assert.Equal(t, "d", e.value.name)
}

func TestParsingTemplateProp2(t *testing.T) {
	str := "abc=def"
	e, err := parseTemplateProp(bufio.NewReader(strings.NewReader(str)))

	assert.Nil(t, err)
	assert.Equal(t, "abc", e.name)
	assert.Equal(t, "def", e.value.name)
}

func TestParsingTemplateProp3(t *testing.T) {
	str := "abc={{def}}"
	e, err := parseTemplateProp(bufio.NewReader(strings.NewReader(str)))

	assert.Nil(t, err)
	assert.Equal(t, "abc", e.name)
	assert.Equal(t, "def", e.value.name)
}

func TestParsingTemplateProp4(t *testing.T) {
	str := "abc={{n|def=doom}}"
	e, err := parseTemplateProp(bufio.NewReader(strings.NewReader(str)))

	assert.Nil(t, err)
	assert.Equal(t, "abc", e.name)
	assert.Equal(t, "n", e.value.name)
	assert.Equal(t, 1, len(e.value.props))
	assert.Equal(t, "def", e.value.props[0].name)
	assert.Equal(t, "doom", e.value.props[0].value.name)
}

func TestParsingTemplate1(t *testing.T) {
	str := "{{abc}}"
	e, err := parseTemplateElement(bufio.NewReader(strings.NewReader(str)))

	assert.Nil(t, err)
	assert.Equal(t, "abc", e.name)
	assert.Equal(t, 0, len(e.props))
}

func TestParsingTemplate2(t *testing.T) {
	str := "{{abc|def}}"
	e, err := parseTemplateElement(bufio.NewReader(strings.NewReader(str)))

	assert.Nil(t, err)
	assert.Equal(t, "abc", e.name)
	assert.Equal(t, 1, len(e.props))
	assert.Equal(t, "def", e.props[0].name)
}

func TestParsingTemplate3(t *testing.T) {
	str := "{{abc|def=boom}}"
	e, err := parseTemplateElement(bufio.NewReader(strings.NewReader(str)))

	assert.Nil(t, err)
	assert.Equal(t, "abc", e.name)
	assert.Equal(t, 1, len(e.props))
	assert.Equal(t, "def", e.props[0].name)
	assert.Equal(t, "boom", e.props[0].value.name)
}

func TestParsingWikiString(t *testing.T) {
	str := "From {{inh|en|enm|breken}}, from {{inh|en|ang|brecan||to break}}, from {{inh|en|gmw-pro|*brekan}}, from {{inh|en|gem-pro|*brekaną||to break}}, from {{inh|en|ine-pro|*bʰreg-||to break}}. The word is a {{doublet|en|bray|nocap=1}}."

	s, err := parseWikiTextString(str)

	assert.Nil(t, err)
	assert.NotNil(t, s)
	assert.Equal(t, WikitextElementTypeText, s.elements[0].elementType())
	assert.Equal(t, WikitextElementTypeTemplate, s.elements[1].elementType())
	assert.Equal(t, WikitextElementTypeText, s.elements[2].elementType())
}

func TestParsingWikitext1(t *testing.T) {
	str := `===Etymology 1===
{{root|en|ine-pro|*bʰreg-}}
From {{inh|en|enm|breken}}, from {{inh|en|ang|brecan||to break}}, from {{inh|en|gmw-pro|*brekan}}, from {{inh|en|gem-pro|*brekaną||to break}}, from {{inh|en|ine-pro|*bʰreg-||to break}}. The word is a {{doublet|en|bray|nocap=1}}.`
	text, err := parseWikitext(str)

	assert.Nil(t, err)
	assert.Equal(t, 3, len(text.strings))
}
