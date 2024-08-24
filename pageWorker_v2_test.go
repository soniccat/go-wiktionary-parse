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

func TestPageWorkerV2_2(t *testing.T) {
	wikitext, err := parseWikitext(TestWikiPage2)
	assert.Nil(t, err)

	inserts := processWikitext("w", wikitext)

	print(InsertsToString(inserts))
	assert.True(t, len(inserts) > 0)
}

func TestPageWorkerV2_3(t *testing.T) {
	wikitext, err := parseWikitext(TestWikiPage3)
	assert.Nil(t, err)

	inserts := processWikitext("w", wikitext)

	print(InsertsToString(inserts))
	assert.True(t, len(inserts) > 0)
}

func TestPageWorkerV2_4(t *testing.T) {
	wikitext, err := parseWikitext(TestWikiPage4)
	assert.Nil(t, err)

	inserts := processWikitext("w", wikitext)

	print(InsertsToString(inserts))
	assert.True(t, len(inserts) > 0)
}
