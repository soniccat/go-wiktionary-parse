package main

func FilterWikitextString(
	wikitext Wikitext,
	stringVisitor func(WikitextString) bool,
	elementVisitor func(WikitextElement) bool,
) Wikitext {
	newWikitext := Wikitext{}

	for _, s := range wikitext.strings {
		if !stringVisitor(s) {
			continue
		}

		var newElements []WikitextElement
		for _, e := range s.elements { // TODO: if necessary need to call visitor for inner elements in template element
			if elementVisitor(e) {
				newElements = append(newElements, e)
			}
		}

		newWikitextString := WikitextString{}
		newWikitextString.elements = newElements
		if !stringVisitor(newWikitextString) {
			continue
		}

		newWikitext.addString(newWikitextString)
	}

	return newWikitext
}

func FilterWikitextMarkup(e WikitextElement) bool {
	_, ok := e.(*WikiMarkupElement)
	return !ok
}

func FilterWikitextEmptyElements(s WikitextString) bool {
	return len(s.elements) > 0
}

func FilterWikitextOrGroup(filters []func(WikitextElement) bool) func(WikitextElement) bool {
	return func(e WikitextElement) bool {
		for _, f := range filters {
			if f(e) {
				return true
			}
		}
		return false
	}
}
