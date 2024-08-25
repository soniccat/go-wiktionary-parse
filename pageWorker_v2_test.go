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

func TestPageWorkerV2_5(t *testing.T) {
	wikitext, err := parseWikitext(TestWikiPage5)
	assert.Nil(t, err)

	inserts := processWikitext("w", wikitext)

	print(InsertsToString(inserts))
	assert.True(t, len(inserts) > 0)
}

func TestPageWorkerV2_6(t *testing.T) {
	wikitext, err := parseWikitext(TestWikiPage6)
	assert.Nil(t, err)

	inserts := processWikitext("w", wikitext)

	print(InsertsToString(inserts))
	assert.True(t, len(inserts) > 0)
}

func TestPageWorkerV2_7(t *testing.T) {
	wikitext, err := parseWikitext(TestWikiPage7)
	assert.Nil(t, err)

	inserts := processWikitext("w", wikitext)

	print(InsertsToString(inserts))
	assert.True(t, len(inserts) > 0)
}

func TestPageWorkerV2_8(t *testing.T) {
	wikitext, err := parseWikitext(TestWikiPage8)
	assert.Nil(t, err)

	inserts := processWikitext("w", wikitext)

	print(InsertsToString(inserts))
	assert.True(t, len(inserts) > 0)
}
