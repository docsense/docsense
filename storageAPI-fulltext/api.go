package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	sc_common "storage-core/common"
	"storage-fulltext/index"
	"storage-fulltext/stemmer"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/fatih/color"
)

type appContext struct {
	stemmer *stemmer.Stemmer
	index   *index.Index
}

type jsonWrapper struct {
	Ctx    *appContext
	Handle func(*appContext, *http.Request) (interface{}, int)
}

func (j jsonWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	obj, status := j.Handle(j.Ctx, r)
	str, err := json.Marshal(obj)
	if err != nil {
		http.Error(w, "could not encode json", 500)
		_, _ = color.New(color.BgMagenta).Printf("JSON ERROR")
		_, _ = color.New(color.FgMagenta).Printf(" " + r.RequestURI + "\n")
		return
	}
	if status >= 400 {
		http.Error(w, string(str), status)
		_, _ = color.New(color.BgRed).Printf("ERROR " + strconv.Itoa(status))
		_, _ = color.New(color.FgRed).Printf(" " + r.RequestURI + "\n")
		return
	}
	_, _ = color.New(color.BgGreen).Printf("OK")
	_, _ = color.New(color.FgGreen).Printf(" " + r.RequestURI + "\n")
	_, _ = w.Write(str) // FIXME NOPE XD
}

type addText struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// AddTextHandler is a handler, that adds text to index
func AddTextHandler(a *appContext, r *http.Request) (interface{}, int) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		sc_common.Error(err)
		return err, 500
	}
	strLen := len(buf.String())
	logrus.Debug("adding text with lenght", strLen)
	decoder := json.NewDecoder(bytes.NewReader(buf.Bytes()))
	var body addText

	err = decoder.Decode(&body)
	if err != nil {
		sc_common.Error(err)
		return err, 400
	}

	//splitted := strings.Split(body.Text, " ")
	splitted := strings.Fields(body.Text)

	words, err := a.stemmer.Stem(splitted)
	if err != nil {
		sc_common.Error(err)
		return err, 400
	}

	a.index.AddDocument(words, index.Document(body.ID))
	logrus.Debug("adding", body.ID)
	return nil, 200
}

type queryText struct {
	Text string `json:"query"`
}

type queryResult struct {
	Result index.ResSlice `json:"result"`
}

// DecodeAndStem parses request and turn slice of words into slice of integers representing
// stems of those words (sans words we deem as useless - "stopwords").
func DecodeAndStem(a *appContext, r *http.Request) ([]int, error) {
	decoder := json.NewDecoder(r.Body)
	var body queryText
	err := decoder.Decode(&body)
	if err != nil {
		return nil, err
	}

	str := strings.Fields(body.Text)
	stemres, err := a.stemmer.Stem(str)

	words := make([]int, 0)

	for _, x := range stemres {
		words = append(words, x.Word)
	}

	if err != nil {
		return nil, err
	}
	return words, nil
}

// PhraseSearchHandler is a handler for searching with some phrase.
func PhraseSearchHandler(a *appContext, r *http.Request) (interface{}, int) {
	words, err := DecodeAndStem(a, r)
	if err != nil {
		return err, 400
	}
	res := a.index.WeightedPhraseSearch(words)
	return queryResult{res}, 200
}

// BowSearchHandler is a handler for searching with mystical BAG-OF-WORDS.
func BowSearchHandler(a *appContext, r *http.Request) (interface{}, int) {
	words, err := DecodeAndStem(a, r)
	if err != nil {
		return err, 400
	}
	res := a.index.BowSearch(words)
	return queryResult{res}, 200
}

// DeprecateDocument deprecates a document which is supposed to be deleted later.
func DeprecateDocument(a *appContext, r *http.Request) (interface{}, int) {
	decoder := json.NewDecoder(r.Body)
	var doc index.Document
	err := decoder.Decode(&doc)
	if err != nil {
		return err, 400
	}
	logrus.Debug("deprecating", doc)
	a.index.Deprecate(doc)
	return nil, 200
}

