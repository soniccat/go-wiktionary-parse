package main

func FilterWikitextString(
	wikitext Wikitext,
	elementVisitor func(WikitextElement) bool,
) Wikitext {
	newWikitext := Wikitext{}

	for _, e := range wikitext.elements { // TODO: if necessary need to call visitor for inner elements in template element
		if elementVisitor(e) {
			newWikitext.addElement(e)
		}
	}

	return newWikitext
}

func FilterWikitextMarkup(e WikitextElement) bool {
	_, ok := e.(*WikitextMarkupElement)
	return !ok
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
