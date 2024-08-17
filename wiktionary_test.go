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

func TestParsingTemplate4(t *testing.T) {
	str := `{{quote-text|en|year=2002|author=w:John Fusco|title={{w|Spirit: Stallion of the Cimarron}}
|passage=Colonel: See, gentlemen? Any horse could be '''broken'''.}}`
	e, err := parseTemplateElement(bufio.NewReader(strings.NewReader(str)))

	assert.Nil(t, err)
	assert.Equal(t, "quote-text", e.name)
	assert.Equal(t, 5, len(e.props))
	assert.Equal(t, "passage", e.props[4].name)
	assert.Equal(t, "Colonel: See, gentlemen? Any horse could be '''broken'''.", e.props[4].value.name)
}

func TestParsingWikitext1(t *testing.T) {
	str := "From {{inh|en|enm|breken}}, from {{inh|en|ang|brecan||to break}}"

	s, err := parseWikitext(str)

	assert.Nil(t, err)
	assert.NotNil(t, s)
	assert.Equal(t, WikitextElementTypeText, s.elements[0].ElementType())
	assert.Equal(t, WikitextElementTypeTemplate, s.elements[1].ElementType())
	assert.Equal(t, WikitextElementTypeText, s.elements[2].ElementType())
	assert.Equal(t, WikitextElementTypeTemplate, s.elements[3].ElementType())

	text, ok := s.elements[0].(*WikitextTextElement)
	assert.True(t, ok)
	assert.Equal(t, "From", text.value)

	template, ok := s.elements[3].(*WikitextTemplateElement)
	assert.True(t, ok)
	assert.Equal(t, "inh", template.name)
}

func TestParsingWikitext2(t *testing.T) {
	str := "#: {{ux|en|If the vase falls to the floor, it might '''break'''.}}"

	s, err := parseWikitext(str)

	assert.Nil(t, err)
	assert.NotNil(t, s)
	assert.Equal(t, 2, len(s.elements))
	assert.Equal(t, WikitextElementTypeMarkup, s.elements[0].ElementType())
	assert.Equal(t, WikitextElementTypeTemplate, s.elements[1].ElementType())

	text, ok := s.elements[0].(*WikitextMarkupElement)
	assert.True(t, ok)
	assert.Equal(t, "#:", text.value)

	template, ok := s.elements[1].(*WikitextTemplateElement)
	assert.True(t, ok)
	assert.Equal(t, "ux", template.name)
}

func TestParsingWikitext3(t *testing.T) {
	str := "# {{lb|en|transitive|intransitive}} To [[separate]] into two or more [[piece]]s, to [[fracture]] or [[crack]], by a process that cannot easily be [[reverse]]d for [[reassembly]]."

	s, err := parseWikitext(str)

	assert.Nil(t, err)
	assert.NotNil(t, s)
	assert.Equal(t, WikitextElementTypeMarkup, s.elements[0].ElementType())
	assert.Equal(t, WikitextElementTypeTemplate, s.elements[1].ElementType())
	assert.Equal(t, WikitextElementTypeText, s.elements[2].ElementType())

	text, ok := s.elements[2].(*WikitextTextElement)
	assert.True(t, ok)
	assert.Equal(
		t,
		"To [[separate]] into two or more [[piece]]s, to [[fracture]] or [[crack]], by a process that cannot easily be [[reverse]]d for [[reassembly]].",
		text.value,
	)
}

func TestParsingWikitext4(t *testing.T) {
	str := `===Etymology 1===
{{root|en|ine-pro|*bʰreg-}}
From {{inh|en|enm|breken}}, from {{inh|en|ang|brecan||to break}}, from {{inh|en|gmw-pro|*brekan}}, from {{inh|en|gem-pro|*brekaną||to break}}, from {{inh|en|ine-pro|*bʰreg-||to break}}. The word is a {{doublet|en|bray|nocap=1}}.`
	text, err := parseWikitext(str)

	assert.Nil(t, err)
	assert.Equal(t, 16, len(text.elements))
	assert.Equal(t, WikitextElementTypeSection, text.elements[0].ElementType())
	assert.Equal(t, WikitextElementTypeTemplate, text.elements[1].ElementType())
}

func TestParsingWikitext5(t *testing.T) {
	str := TestWikiPage
	text, err := parseWikitext(str)

	assert.Nil(t, err)
	assert.True(t, len(text.elements) > 0)
}

func TestParsingWikitext6(t *testing.T) {
	str := `#: {{ux|en|If the vase falls to the floor, it might '''break'''.}}
#: {{ux|en|In order to tend to the accident victim, he will '''break''' the window of the car.}}`
	text, err := parseWikitext(str)

	assert.Nil(t, err)
	assert.Equal(t, 5, len(text.elements))
	assert.Equal(t, WikitextElementTypeMarkup, text.elements[0].ElementType())
	assert.Equal(t, WikitextElementTypeTemplate, text.elements[1].ElementType())
	assert.Equal(t, WikitextElementTypeNewline, text.elements[2].ElementType())
	assert.Equal(t, WikitextElementTypeMarkup, text.elements[3].ElementType())
	assert.Equal(t, WikitextElementTypeTemplate, text.elements[4].ElementType())
}

func TestParsingWikitext7(t *testing.T) {
	str := `# {{lb|en|transitive|intransitive}} To [[separate]] into two or more [[piece]]s, to [[fracture]] or [[crack]], by a process that cannot easily be [[reverse]]d for [[reassembly]].
#: {{ux|en|If the vase falls to the floor, it might '''break'''.}}`
	text, err := parseWikitext(str)

	assert.Nil(t, err)
	assert.Equal(t, 6, len(text.elements))
	assert.Equal(t, WikitextElementTypeMarkup, text.elements[0].ElementType())
	assert.Equal(t, WikitextElementTypeTemplate, text.elements[1].ElementType())
	assert.Equal(t, WikitextElementTypeText, text.elements[2].ElementType())
	assert.Equal(t, WikitextElementTypeNewline, text.elements[3].ElementType())
	assert.Equal(t, WikitextElementTypeMarkup, text.elements[4].ElementType())
	assert.Equal(t, WikitextElementTypeTemplate, text.elements[5].ElementType())
}
