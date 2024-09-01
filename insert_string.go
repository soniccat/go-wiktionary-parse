package main

import (
	"fmt"
	"strings"
)

func InsertsToString(inserts []WordEntry) string {
	strBuilder := strings.Builder{}

	for _, insert := range inserts {
		strBuilder.WriteString(fmt.Sprintf("Term: %s\n", insert.Term))
		strBuilder.WriteString("Transcriptions:\n")
		for _, t := range insert.Transcriptions {
			strBuilder.WriteString(t + ", ")
		}
		strBuilder.WriteString("\n")

		for _, v := range insert.DefPairs {
			strBuilder.WriteString(v.PartOfSpeech + "\n")

			for _, d := range v.DefEntries {
				strBuilder.WriteString("d: " + d.Def.Value + "\n")

				if len(d.Def.Labels) > 0 {
					strBuilder.WriteString("l: ")
					for _, l := range d.Def.Labels {
						strBuilder.WriteString(l + ", ")
					}
					strBuilder.WriteString("\n")
				}

				for _, e := range d.Examples {
					strBuilder.WriteString("\te: " + e + "\n")
				}
				for _, s := range d.Synonyms {
					strBuilder.WriteString("\ts: " + s + "\n")
				}
				for _, a := range d.Antonyms {
					strBuilder.WriteString("\ta: " + a + "\n")
				}
			}
		}
	}

	return strBuilder.String()
}
