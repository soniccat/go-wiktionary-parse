package main

import (
	"database/sql"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
)

type WordEntry struct {
	Order          int                       //`bson:"order"`
	Word           string                    `bson:"term,omitempty"`
	Transcriptions []string                  `bson:"transcriptions,omitempty"`
	Etymology      int                       `bson:"etymology,omitempty"`
	WordDefs       map[string][]WordDefEntry `bson:"defs,omitempty"`
}

type WordDefEntry struct {
	WordDef  WordDef  `bson:"def,omitempty"`
	Examples []string `bson:"examples,omitempty"`
	Synonyms []string `bson:"synonyms,omitempty"`
	Antonyms []string `bson:"antonyms,omitempty"`
}

type WordDef struct {
	Def    string   `bson:"def,omitempty"`
	Labels []string `bson:"labels,omitempty"`
}

func pageWorkerV2(
	id int,
	wg *sync.WaitGroup,
	pages []Page,
	dbh *sql.DB,
	mongo *mongo.Collection,
) []WordEntry {
	defer wg.Done()
	inserts := []WordEntry{}
	for _, page := range pages {
		word := page.Title
		logger.Debug("Processing page: %s\n", word)
		logger.Debug("text: %s\n", page.Revisions[0].Text)

		w, err := parseWikitext(page.Revisions[0].Text)
		if err != nil {
			logger.Error("parse error for %s, %v", page.Title, err.Error())
			logger.Error("text %s", page.Revisions[0].Text)
			continue
		}

		// w = FilterWikitextString(
		// 	w,
		// 	FilterWikitextMarkup,
		// )

		inserts = append(inserts, processWikitext(word, w)...)
	}

	return inserts

	// // perform inserts
	// inserted := performInserts(dbh, inserts)
	// if mongo != nil {
	// 	documents := make([]interface{}, len(inserts))
	// 	for i := range inserts {
	// 		documents[i] = inserts[i]
	// 	}
	// 	r, err := mongo.InsertMany(context.Background(), documents)
	// 	logger.Debug("%v %v", r, err)
	// }
	// logger.Info("[%2d] Inserted %6d records for %6d pages\n", id, inserted, len(pages))
}

func processWikitext(word string, wikitext Wikitext) []WordEntry {
	cb := CardBuilder{}
	cb.SetWord(word)

	inPartOfSpeech := false
	languageSectionLevel := -1
	areSynonyms := false
	areAntonyms := false
	isDefinition := false
	isExample := false

	// read elements until English language section
	var elementIndex int
	for i, e := range wikitext.elements {
		elementIndex = i
		section, ok := e.(*WikitextSectionElement)
		if ok && section.name == "English" {
			languageSectionLevel = section.level
			break
		}
	}

	var textElements []string
	var labels []string

	for _, e := range wikitext.elements[elementIndex+1:] {
		switch re := e.(type) {
		case *WikitextSectionElement:
			areSynonyms = false
			areAntonyms = false
			inPartOfSpeech = false
			if re.level < languageSectionLevel {
				break
			} else if strings.HasPrefix(re.name, "Etymology") {
				cb.StartEtymology()
			} else if strings.HasPrefix(re.name, "Synonyms") {
				areSynonyms = true
			} else if strings.HasPrefix(re.name, "Antonyms") {
				areAntonyms = true
			}

		case *WikitextTemplateElement:
			switch re.name {
			case "enPR", "IPA":
				offset := 0
				if re.name == "IPA" {
					offset = 1
				}
				for _, v := range re.props[offset:] {
					if v.isStringValue() {
						cb.AddTranscription(v.stringValue())
					}
				}
			case "en-verb":
				inPartOfSpeech = true
				cb.SetPartOfSpeech("verb")
			case "en-noun":
				inPartOfSpeech = true
				cb.SetPartOfSpeech("noun")
			case "en-adj":
				inPartOfSpeech = true
				cb.SetPartOfSpeech("adj")
			case "en-adv":
				inPartOfSpeech = true
				cb.SetPartOfSpeech("adv")
			case "en-con":
				inPartOfSpeech = true
				cb.SetPartOfSpeech("con")
			case "en-det":
				inPartOfSpeech = true
				cb.SetPartOfSpeech("det")
			case "en-interj":
				inPartOfSpeech = true
				cb.SetPartOfSpeech("interj")
			case "en-num":
				inPartOfSpeech = true
				cb.SetPartOfSpeech("num")
			case "en-part":
				inPartOfSpeech = true
				cb.SetPartOfSpeech("part")
			case "en-postp":
				inPartOfSpeech = true
				cb.SetPartOfSpeech("postp")
			case "en-prep":
				inPartOfSpeech = true
				cb.SetPartOfSpeech("prep")
			case "en-pron":
				inPartOfSpeech = true
				cb.SetPartOfSpeech("pron")
			case "en-proper noun":
				inPartOfSpeech = true
				cb.SetPartOfSpeech("proper noun")
			case "lb":
				for i, v := range re.props {
					if i > 0 && v.isStringValue() {
						labels = append(labels, v.stringValue())
					}
				}
			case "syn", "synonyms":
				for i, v := range re.props {
					if i > 0 && v.isStringValue() {
						cb.AddSynonym(v.stringValue())
					}
				}
			case "ux":
				if len(re.props) > 1 && re.props[1].isStringValue() {
					cb.AddExample(re.props[1].stringValue())
				}
			case "quote-book", "quote-text", "quote-av", "quote-hansard", "quote-journal", "quote-mailing list", "quote-newsgroup", "quote-song", "quote-us-patent", "quote-video game", "quote-web", "quote-wikipedia":
				if inPartOfSpeech {
					textProp := re.PropByName("text")
					if textProp == nil {
						textProp = re.PropByName("passage")
					}

					var ex string
					if textProp != nil && textProp.isInnerStringValue() {
						ex = textProp.innerStringValue()
					} else {
						if re.name == "quote-book" && textProp == nil {
							seventhProp := re.PropStringPropByIndex(6)
							if seventhProp != nil && seventhProp.isStringValue() {
								ex = seventhProp.stringValue()
							}
						} else if re.name == "quote-journal" && textProp == nil {
							eigthProp := re.PropStringPropByIndex(7)
							if eigthProp != nil && eigthProp.isStringValue() {
								ex = eigthProp.stringValue()
							}
						}
					}

					if len(ex) > 0 {
						cb.AddExample(ex)
					}
				}
			case "sense":
				if (areSynonyms || areAntonyms) && len(re.props) > 0 && re.props[0].isStringValue() {
					cb.AddDefinition(re.props[0].stringValue(), nil)
				}
			case "l":
				if (areSynonyms || areAntonyms) && len(re.props) > 1 && re.props[1].isStringValue() {
					if areSynonyms {
						cb.AddSynonym(re.props[1].stringValue())
					} else if areAntonyms {
						cb.AddAntonym(re.props[1].stringValue())
					}
				}
			}

		case *WikitextMarkupElement:
			if inPartOfSpeech && strings.HasSuffix(re.value, "#") {
				isDefinition = true
			} else if inPartOfSpeech && strings.HasSuffix(re.value, ":") {
				isExample = true
			}

		case *WikitextTextElement:
			if inPartOfSpeech {
				textElements = append(textElements, re.value)
			}

		case *WikitextNewlineElement:
			if isDefinition && len(textElements) > 0 {
				d := strings.Join(textElements, " ")
				cb.AddDefinition(d, labels)
			} else if isExample && len(textElements) > 0 {
				ex := strings.Join(textElements, " ")
				cb.AddExample(ex)
			}

			textElements = nil
			labels = nil
			isDefinition = false
			isExample = false
		}
	}

	return cb.Build()
}

type CardBuilder struct {
	isEtymologyStarted   bool
	globalTranscriptions []string
	currentInsert        WordEntry
	currentPartOfSpeech  string
	currentDef           WordDefEntry
	inserts              []WordEntry
}

func (cb *CardBuilder) SetWord(w string) {
	cb.currentInsert.Word = w
	cb.currentInsert.WordDefs = make(map[string][]WordDefEntry)
}

func (cb *CardBuilder) AddTranscription(t string) {
	if cb.isEtymologyStarted {
		cb.currentInsert.Transcriptions = append(cb.currentInsert.Transcriptions, t)
	} else {
		cb.globalTranscriptions = append(cb.globalTranscriptions, t)
	}
}

func (cb *CardBuilder) StartEtymology() {
	cb.save()
	cb.isEtymologyStarted = true
}

func (cb *CardBuilder) SetPartOfSpeech(s string) {
	//cb.save()
	cb.currentPartOfSpeech = s
}

func (cb *CardBuilder) AddDefinition(d string, labels []string) {
	cb.saveDefinition()
	cb.currentDef.WordDef = WordDef{
		Def:    d,
		Labels: labels,
	}
}

func (cb *CardBuilder) AddExample(e string) {
	cb.currentDef.Examples = append(cb.currentDef.Examples, e)
}

func (cb *CardBuilder) AddSynonym(s string) {
	cb.currentDef.Synonyms = append(cb.currentDef.Synonyms, s)
}

func (cb *CardBuilder) AddAntonym(a string) {
	cb.currentDef.Antonyms = append(cb.currentDef.Antonyms, a)
}

// TODO: support hyponym / Derived terms

func (cb *CardBuilder) save() {
	if len(cb.currentInsert.WordDefs) > 0 {
		if len(cb.currentInsert.Transcriptions) == 0 {
			cb.currentInsert.Transcriptions = append(cb.currentInsert.Transcriptions, cb.globalTranscriptions...)
		}

		cb.inserts = append(cb.inserts, cb.currentInsert)
	}

	cb.currentInsert = WordEntry{
		Word:      cb.currentInsert.Word,
		Etymology: len(cb.inserts),
		WordDefs:  make(map[string][]WordDefEntry),
	}
}

func (cb *CardBuilder) saveDefinition() {
	if len(cb.currentDef.WordDef.Def) > 0 {
		defs := cb.currentInsert.WordDefs[cb.currentPartOfSpeech]
		cb.currentInsert.WordDefs[cb.currentPartOfSpeech] = append(defs, cb.currentDef)
	}

	cb.currentDef = WordDefEntry{}
}

func (cb *CardBuilder) Build() []WordEntry {
	cb.saveDefinition()
	cb.save()
	return cb.inserts
}
