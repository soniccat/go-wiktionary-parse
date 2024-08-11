package main

import (
	"context"
	"database/sql"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
)

func pageWorkerV2(
	id int,
	wg *sync.WaitGroup,
	pages []Page,
	dbh *sql.DB,
	mongo *mongo.Collection,
) {
	defer wg.Done()
	inserts := []*Insert{} // etymology : lexical category : [definitions...]
	for _, page := range pages {
		word := page.Title
		logger.Debug("Processing page: %s\n", word)

		w, err := parseWikitext(page.Revisions[0].Text)
		if err != nil {
			logger.Error("parse error for %s, %v", page.Title, err.Error())
			continue
		}

		transcriptions []
	}

	// perform inserts
	inserted := performInserts(dbh, inserts)
	if mongo != nil {
		documents := make([]interface{}, len(inserts))
		for i := range inserts {
			documents[i] = inserts[i]
		}
		r, err := mongo.InsertMany(context.Background(), documents)
		logger.Debug("%v %v", r, err)
	}

	logger.Info("[%2d] Inserted %6d records for %6d pages\n", id, inserted, len(pages))
}


type CardBuilder struct {
	
}