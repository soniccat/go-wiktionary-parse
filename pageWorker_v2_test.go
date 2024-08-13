package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPageWorkerV2_1(t *testing.T) {
	wikitext, err := parseWikitext(TestWikiPage)
	assert.Nil(t, err)

	inserts := processWikitext("w", wikitext)

	print(InsertsToString(inserts))
	assert.True(t, len(inserts) > 0)
}
