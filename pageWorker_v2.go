package main

import (
	"database/sql"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
)

func pageWorkerV2(
	id int,
	wg *sync.WaitGroup,
	pages []Page,
	dbh *sql.DB,
	mongo *mongo.Collection,
) []Insert {
	defer wg.Done()
	inserts := []Insert{} // etymology : lexical category : [definitions...]
	for _, page := range pages {
		word := page.Title
		logger.Debug("Processing page: %s\n", word)

		w, err := parseWikitext(page.Revisions[0].Text)
		if err != nil {
			logger.Error("parse error for %s, %v", page.Title, err.Error())
			continue
		}

		w = FilterWikitextString(
			w,
			FilterWikitextMarkup,
		)

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

func processWikitext(word string, wikitext Wikitext) []Insert {
	cb := CardBuilder{}
	cb.SetWord(word)

	// scanUntiEnglish := true
	// scanUntiEtymology := false
	inPartOfSpeech := false
	languageSectionLevel := -1

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
			if re.level < languageSectionLevel {
				break
			} else if strings.HasPrefix(re.name, "Etymology") {
				// scanUntiEtymology = false
				cb.StartEtymology()
			} else {
				inPartOfSpeech = false
			}
		case *WikiTemplateElement:
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

			case "lb":
				for i, v := range re.props {
					if i > 0 && v.isStringValue() {
						labels = append(labels, v.stringValue())
					}
				}

			case "ux":
				if inPartOfSpeech && len(textElements) > 0 {
					d := strings.Join(textElements, " ")
					cb.AddDefinition(d, labels)
					textElements = nil
					labels = nil
				}

				if re.name == "ux" {
					if len(re.props) > 1 && re.props[1].isStringValue() {
						cb.AddExample(re.props[1].stringValue())
					}
				}
			}
		case *WikiTextElement:
			if inPartOfSpeech {
				textElements = append(textElements, re.value)
			}
		}
	}

	return cb.Build()
}

type CardBuilder struct {
	isEtymologyStarted   bool
	globalTranscriptions []string
	currentInsert        Insert
	currentPartOfSpeech  string
	currentDef           CatDef
	inserts              []Insert
}

func (cb *CardBuilder) SetWord(w string) {
	cb.currentInsert.Word = w
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

func (cb *CardBuilder) AddDefinition(d string, labels []string) { // TODO: store labels
	cb.saveDefinition()
	cb.currentDef.Def = d
}

func (cb *CardBuilder) AddExample(e string) {
	cb.currentDef.Examples = append(cb.currentDef.Examples, e)
}

func (cb *CardBuilder) AddSynonym(s string, labels []string) { // TODO: store labels
	cb.currentDef.Synonyms = append(cb.currentDef.Synonyms, s)
}

func (cb *CardBuilder) AddAntonym(a string, labels []string) { // TODO: store labels
	cb.currentDef.Antonyms = append(cb.currentDef.Antonyms, a)
}

// TODO: support hyponym / Derived terms

func (cb *CardBuilder) save() {
	if len(cb.currentInsert.CatDefs) > 0 {
		if len(cb.currentInsert.Transcriptions) == 0 {
			cb.currentInsert.Transcriptions = append(cb.currentInsert.Transcriptions, cb.globalTranscriptions...)
		}

		cb.inserts = append(cb.inserts, cb.currentInsert)
	}

	cb.currentInsert = Insert{
		Word:      cb.currentInsert.Word,
		Etymology: len(cb.inserts),
		CatDefs:   make(map[string][]CatDef),
	}
}

func (cb *CardBuilder) saveDefinition() {
	if len(cb.currentDef.Def) > 0 {
		defs := cb.currentInsert.CatDefs[cb.currentPartOfSpeech]
		cb.currentInsert.CatDefs[cb.currentPartOfSpeech] = append(defs, cb.currentDef)
	}

	cb.currentDef = CatDef{}
}

func (cb *CardBuilder) Build() []Insert {
	cb.save()
	return cb.inserts
}
