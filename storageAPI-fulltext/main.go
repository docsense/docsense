package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	sc_common "storage-core/common"
	"storage-fulltext/index"
	"storage-fulltext/stemmer"
	"time"

	"github.com/Sirupsen/logrus"
)

func main() {
	stm := stemmer.NewStemmer()
	idx := index.Create()

	context := &appContext{
		stemmer: stm,
		index:   idx,
	}

	ticker := time.NewTicker(10 * time.Minute)

	idx.PrintMeta()
	go tick(ticker, idx)
	defer ticker.Stop()

	r := mux.NewRouter()
	r.Handle("/add", jsonWrapper{context, AddTextHandler}).Methods("POST")
	r.Handle("/search/phrase", jsonWrapper{context, PhraseSearchHandler}).Methods("POST")
	r.Handle("/search/bow", jsonWrapper{context, BowSearchHandler}).Methods("POST")
	r.Handle("/deprecate", jsonWrapper{context, DeprecateDocument}).Methods("POST")
	logrus.Info("starting application!")

	s := &http.Server{
		Addr:           ":8888",
		Handler:        r,
		ReadTimeout:    1000 * time.Second,
		WriteTimeout:   1000 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := s.ListenAndServe()
	if err != nil {
		sc_common.Error(err)
	}
}

func tick(ticker *time.Ticker, idx *index.Index) {
	for t := range ticker.C {
		fmt.Print(t.Format("2006-01-02 15:04:05") + " ")
		idx.PrintMeta()
	}
}