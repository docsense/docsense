package fulltext

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"storage-core/common"
	"strconv"
	"time"
)

//Fulltext represents a fulltext search instance
type Fulltext struct {
	Addr string
	Port string
}

func path(f *Fulltext) string {
	return "http://" + f.Addr + ":" + f.Port + "/"
}

type addJSON struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

//Create creates a new storage fulltext instance
func Create(config common.Config) Fulltext {
	ft := new(Fulltext)
	ft.Addr = config.Fulltext.Addr
	ft.Port = config.Fulltext.Port
	return *ft
}

//AddText adds a text to storage fulltext. r must be a text file.
func (f *Fulltext) AddText(file common.File, r io.Reader) error {
	qjson := addJSON{ID: file.FullQualifier(), Text: common.StreamToString(r)}
	str, err := json.Marshal(qjson)
	if err != nil {
		return err
	}

	var NetClient = http.Client{
		Timeout: time.Second * 30,
		Transport: &http.Transport{
			DisableKeepAlives:   true,
			MaxIdleConnsPerHost: 100,
		},
	}

	req, err := http.NewRequest("POST", path(f)+"add", bytes.NewReader(str))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	res, err := NetClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return errors.New("Status Error" + strconv.Itoa(res.StatusCode))
	}
	return nil
}

type queryJSON struct {
	Text string `json:"query"`
}

//Response represents a storage fulltext response
type Response struct {
	Result []struct {
		Doc    string `json:"Doc"`
		Score  int    `json:"Score"`
		PosBeg int    `json:"PosBeg"`
		PosEnd int    `json:"PosEnd"`
	} `json:"Result"`
}

func (f *Fulltext) doQuery(s, method string) (*Response, error) {
	qjson := queryJSON{s}
	str, err := json.Marshal(qjson)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(path(f)+"search/"+method, "text/plain", bytes.NewReader(str))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error - status code %d", resp.StatusCode)
	}
	var result Response

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

//PhraseQuery does a fulltext phrase query (a few words)
func (f *Fulltext) PhraseQuery(s string) (*Response, error) {
	return f.doQuery(s, "phrase")
}

//BowQuery does a fulltext bow query (a whole document)
func (f *Fulltext) BowQuery(s string) (*Response, error) {
	return f.doQuery(s, "bow")
}

//DeprecateDocument removes a document from the fulltext index
//TODO pass common.File instead of a string
func (f *Fulltext) DeprecateDocument(file string) error {
	str, err := json.Marshal(file)
	if err != nil {
		return err
	}
	response, err := http.Post(path(f)+"deprecate", "text/plain", bytes.NewReader(str))
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("Error - status code %s", response.Status)
	}
	return nil
}
