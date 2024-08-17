package main

import (
	"fmt"
	"strings"
)

func InsertsToString(inserts []WordEntry) string {
	strBuilder := strings.Builder{}

	for _, insert := range inserts {
		strBuilder.WriteString(fmt.Sprintf("Term: %s\n", insert.Word))
		strBuilder.WriteString("Transcriptions:\n")
		for _, t := range insert.Transcriptions {
			strBuilder.WriteString(t + ", ")
		}
		strBuilder.WriteString("\n")

		for k, t := range insert.WordDefs {
			strBuilder.WriteString(k + "\n")

			for _, d := range t {
				strBuilder.WriteString("d: " + d.WordDef.Def + "\n")

				if len(d.WordDef.Labels) > 0 {
					strBuilder.WriteString("l: ")
					for _, l := range d.WordDef.Labels {
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
