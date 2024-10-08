package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsingSection1(t *testing.T) {
	str := "=t="
	e, err := parseSectionElement(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "t", e.name)
	assert.Equal(t, 1, e.level)
}

func TestParsingSection2(t *testing.T) {
	str := "==ttt=="
	e, err := parseSectionElement(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "ttt", e.name)
	assert.Equal(t, 2, e.level)
}

func TestParsingTemplateProp0(t *testing.T) {
	str := "a"
	e, err := parseTemplateProp(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "a", e.name)
	assert.Nil(t, e.value)
}

func TestParsingTemplateProp1(t *testing.T) {
	str := "a=d"
	e, err := parseTemplateProp(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "a", e.name)
	assert.Equal(t, "d", *e.value)
}

func TestParsingTemplateProp2(t *testing.T) {
	str := "abc=def"
	e, err := parseTemplateProp(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "abc", e.name)
	assert.Equal(t, "def", *e.value)
}

func TestParsingTemplateProp3(t *testing.T) {
	str := "abc={{def}}"
	e, err := parseTemplateProp(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "abc", e.name)
	assert.Equal(t, "def", *e.value)
}

func TestParsingTemplateProp4(t *testing.T) {
	str := "abc={{n|def=doom}}"
	e, err := parseTemplateProp(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "abc", e.name)
	assert.Equal(t, "doom", *e.value)
	// assert.Equal(t, 1, len(e.value.props))
	// assert.Equal(t, "def", e.value.props[0].name)
	// assert.Equal(t, "doom", e.value.props[0].value.name)
}

func TestParsingTemplateProp5(t *testing.T) {
	str := "w:[[Rail (magazine)|Rail]]"
	e, err := parseTemplateProp(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "w:Rail", e.name)
	assert.Nil(t, e.value)
}
func TestParsingTemplateProp6(t *testing.T) {
	str := "w:[[Rail (magazine)]]"
	e, err := parseTemplateProp(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "w:Rail (magazine)", e.name)
	assert.Nil(t, e.value)
}

func TestParsingTemplateProp7(t *testing.T) {
	str := "passage={{...}} the '''hypermasculinized''' image of rappers such as Puff Daddy (Sean Combs) {{...}}"
	e, err := parseTemplateProp(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "passage", e.name)
	assert.Equal(t, "... the hypermasculinized image of rappers such as Puff Daddy (Sean Combs) ...", e.innerStringValue())
}

func TestParsingTemplateProp8(t *testing.T) {
	str := "title={{w|Taming of the Shrew}}, I, ii"
	e, err := parseTemplateProp(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "title", e.name)
	assert.Equal(t, "Taming of the Shrew, I, ii", e.innerStringValue())
}

func TestParsingTemplateProp9(t *testing.T) {
	str := "passage=The redshift of light leaking outward from the '''photon sphere''' is <math>\\sqrt{3} - 1 = 0.732</math>. All light rays approaching a black hole closer than <math>\\sqrt{3}</math> times the radius of the '''photon sphere''' spiral inwards and are captured (see Figure 13.5)."
	e, err := parseTemplateProp(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "passage", e.name)
	assert.Equal(t, "The redshift of light leaking outward from the photon sphere is \\sqrt{3} - 1 = 0.732. All light rays approaching a black hole closer than \\sqrt{3} times the radius of the photon sphere spiral inwards and are captured (see Figure 13.5).", e.innerStringValue())
}

func TestParsingTemplateProp10(t *testing.T) {
	str := "passage=''Heinrich Olbers described the paradox that '''bears''' his name in 1823.''"
	e, err := parseTemplateProp(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "passage", e.name)
	assert.Equal(t, "Heinrich Olbers described the paradox that bears his name in 1823.", e.innerStringValue())
}

func TestParsingTemplateProp11(t *testing.T) {
	str := "passage=I vote we go down and explore. No one will come up this time o' day. We needn't '''keep ''cavé'''''."
	e, err := parseTemplateProp(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "passage", e.name)
	assert.Equal(t, "I vote we go down and explore. No one will come up this time o' day. We needn't keep cavé.", e.innerStringValue())
}

func TestParsingTemplate1(t *testing.T) {
	str := "{{abc}}"
	e, err := parseTemplateElement(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "abc", e.name)
	assert.Equal(t, 0, len(e.props))
}

func TestParsingTemplate2(t *testing.T) {
	str := "{{abc|def}}"
	e, err := parseTemplateElement(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "abc", e.name)
	assert.Equal(t, 1, len(e.props))
	assert.Equal(t, "def", e.props[0].name)
}

func TestParsingTemplate3(t *testing.T) {
	str := "{{abc|def=boom}}"
	e, err := parseTemplateElement(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "abc", e.name)
	assert.Equal(t, 1, len(e.props))
	assert.Equal(t, "def", e.props[0].name)
	assert.Equal(t, "boom", *e.props[0].value)
}

func TestParsingTemplate4(t *testing.T) {
	str := `{{quote-text|en|year=2002|author=w:John Fusco|title={{w|Spirit: Stallion of the Cimarron}}
|passage=Colonel: See, gentlemen? Any horse could be '''broken'''.}}`
	e, err := parseTemplateElement(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "quote-text", e.name)
	assert.Equal(t, 5, len(e.props))
	assert.Equal(t, "passage", e.props[4].name)
	assert.Equal(t, "Colonel: See, gentlemen? Any horse could be broken.", *e.props[4].value)
}

func TestParsingTemplate5(t *testing.T) {
	str := "{{abc|def|_|puf|duf}}"
	e, err := parseTemplateElement(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "abc", e.name)
	assert.Equal(t, 2, len(e.props))
	assert.Equal(t, "def puf", e.props[0].name)
	assert.Equal(t, "duf", e.props[1].name)
}

func TestParsingTemplate6(t *testing.T) {
	str := "{{abc|def|or|puf}}"
	e, err := parseTemplateElement(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "abc", e.name)
	assert.Equal(t, 1, len(e.props))
	assert.Equal(t, "def or puf", e.props[0].name)
}

func TestParsingTemplate7(t *testing.T) {
	str := "{{abc|def|and|puf}}"
	e, err := parseTemplateElement(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "abc", e.name)
	assert.Equal(t, 1, len(e.props))
	assert.Equal(t, "def and puf", e.props[0].name)
}

func TestParsingTemplate8(t *testing.T) {
	str := "{{abc|def|;|puf}}"
	e, err := parseTemplateElement(strings.NewReader(str))

	assert.Nil(t, err)
	assert.Equal(t, "abc", e.name)
	assert.Equal(t, 1, len(e.props))
	assert.Equal(t, "def; puf", e.props[0].name)
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
	str := "# {{lb|en|transitive|intransitive}} To [[separate]] into '''two''' or more [[piece]]s, to [[fracture]] or [[crack]], by a process that cannot easily be [[reverse]]d for [[reassembly]]."

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
		"To separate into two or more pieces, to fracture or crack, by a process that cannot easily be reversed for reassembly.",
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

func TestParsingWikitext8(t *testing.T) {
	str := `==English==

===Etymology===
{{prefix|en|hyper|masculinized}}

===Adjective===
{{en-adj}}

# Extremely [[masculinize]]d.
#* {{quote-text|en|year=2006|author=Robert C. Smith|title=Mexican New York: transnational lives of new immigrants|page=132|passage={{...}} the '''hypermasculinized''' image of rappers such as Puff Daddy (Sean Combs) {{...}}}}`
	text, err := parseWikitext(str)

	assert.Nil(t, err)
	assert.Equal(t, 15, len(text.elements))
}

func TestParsingWikitext9(t *testing.T) {
	str := `==English==

===Noun===
{{en-noun|~}}

# {{lb|en|organic compound}} The [[alkaloid]] ''(2S,3R,11bS)-3-ethyl-2-[[(1R)-2,3,4,9-tetrahydro-1H-pyrido[3,4-b]indol-1-yl]methyl]-2,3,4,6,7,11b-hexahydro-1H-benzo[a]quinolizine''`
	text, err := parseWikitext(str)

	assert.Nil(t, err)
	assert.Equal(t, 9, len(text.elements))
	te := text.elements[8].(*WikitextTextElement)
	assert.Equal(t, `The alkaloid (2S,3R,11bS)-3-ethyl-2-[[(1R)-2,3,4,9-tetrahydro-1H-pyrido[3,4-b]indol-1-yl]methyl]-2,3,4,6,7,11b-hexahydro-1H-benzo[a]quinolizine`, te.value)
}

func TestParsingWikitext10(t *testing.T) {
	str := `==English==

===Adjective===
{{en-adj}}

# {{alternative form of|en|Afrophobic}}
#* {{quote-book|en|passage=Those reactions, whether genuine or effected, were far less about Black Muslims and far more about black people in their perpetual role as the emotional babysitter of the '''''Afriphobic''''' white community in the United States.|page=159|author=Louis A. DeCaro, Jr.|title=Malcolm and the Cross: The Nation of Islam, Malcolm X, and Christianity|publisher=New York University Press|year=1998|url=https://www.google.co.uk/books/edition/Malcolm_and_the_Cross/sSYTCgAAQBAJ}}
`
	text, err := parseWikitext(str)
	assert.Equal(t, 12, len(text.elements))
	assert.Nil(t, err)
}

func TestParsingWikitext11(t *testing.T) {
	str := `==English==

===Etymology===
Used at {{w|Eton College}}, [[Berkshire]], [[England]]. See {{m|en|cave|t=look out, beware}}.

===Pronunciation===
* {{enPR|kēp kāʹvē}}, {{IPA|en|/kiːp ˈkeɪvi/}}

===Verb===
{{en-verb|keep<,,kept> cave}}

# {{lb|en|Britain|public school slang}} To [[maintain]] [[vigilance]].
#: {{synonyms|en|keep a lookout|keep watch}}
#* {{RQ:Kipling Stalky|page=73|passage=I vote we go down and explore. No one will come up this time o' day. We needn't '''keep ''cavé'''''.}}
#* {{quote-book|en|author=Cosmo Hamilton|chapter=Why Cupid Came to Earl’s Court|url=http://archive.org/details/cu31924013622950|title=Short Plays for Small Stages|year=1911|page=9|passage=I'll '''keep cave''' for ten minutes.}}
#* {{quote-book|en|author=Ethel Turner|title=Seven Little Australians|url=https://archive.org/details/sevenlittleaustr00turniala|year=1912|page=154|passage=Pip was mounting guard at the shed, and had undertaken to get Judy safely away, and Bunty had been stationed on the back verandah to '''keep ''cave''''' and whistle three times if there was any danger.}}
`
	text, err := parseWikitext(str)
	assert.Equal(t, 37, len(text.elements))
	assert.Nil(t, err)
}

func TestParsingWikitext12(t *testing.T) {
	str := `#* <!--original work is...?-->'''1905''', Emily Wilder Leavitt, ''Palmer Groups: John Melvin of Charlestown and Concord, Mass. and His Descendants ; Gathered and Arranged for Mr. Lowell Mason Palmer of New York'', page 24:`
	text, err := parseWikitext(str)
	assert.Equal(t, 2, len(text.elements))
	assert.Nil(t, err)
}

func TestParsingWikitext13(t *testing.T) {
	str := `
==English==

===Etymology 1===

====Pronunciation====
* {{enPR|bâr}}

====Noun====
{{en-noun|-}}

# {{alternative spelling of|en|bere||barley}}.
#* '''1800''', Tuke, ''Agric.'', 119:
#*: There are several plots of those species of barley called big, which is six-rowed barley; or '''bear''', which is four-rowed, cultivated.
#* {{quote-text|en|year=1818|location=Marshall|title=Reports Agric.|section=I. 191
|passage=Bigg or bear, with four grains on the ear, was the kind of barley.}}
`
	text, err := parseWikitext(str)

	inserts := processWikitext("w", text)

	assert.Equal(t, 1, len(inserts))
	assert.Nil(t, err)
}

func TestParsingWikitext14(t *testing.T) {
	str := `
==English==

===Etymology 1===

====Pronunciation====
* {{enPR|bâr}}

====Noun====
{{en-noun|-}}


# {{lb|en|sports|and|games}}:
## {{lb|en|transitive|tennis}} To [[win]] a [[game]] (against one's [[opponent]]) as [[receiver]].
`
	text, err := parseWikitext(str)

	inserts := processWikitext("w", text)

	assert.Equal(t, 1, len(inserts))
	assert.Nil(t, err)
}
