package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"storage-core/common"
	"storage-core/disk"
	"strconv"
	"strings"

	"gopkg.in/mgo.v2/bson"

	"github.com/fatih/color"
	//    "bytes"

	"storage-core/mongo"
	"storage-core/search/fulltext"

	"github.com/Sirupsen/logrus"
)

// Zobacz https://elithrar.github.io/article/custom-handlers-avoiding-globals/

type appContext struct {
	storage  disk.APIConnection
	fulltext *fulltext.Fulltext
	mongo    *mongo.Mongo
}

type jsonWrapper struct {
	Ctx    *appContext
	Handle func(*appContext, common.App, *http.Request) (interface{}, int)
}

type fileWrapper struct {
	Ctx    *appContext
	Handle func(*appContext, common.App, *http.Request) (io.Reader, int)
}

var tokens = make(map[string]common.App)
var maxFile = 0

func initTokens(config common.Config) {
	for _, app := range config.Apps {
		tokens[app.Token] = app
	}
}

func nextNumString(a *appContext) (string, error) {
	ret := strconv.Itoa(maxFile)
	maxFile++
	err := a.mongo.PutTuple("maxdocument", bson.M{"maxdocument": maxFile})
	return ret, err
}

var validMetadata = regexp.MustCompile(`^([ żółćęśąźńŻÓŁĆĘŚĄŹŃa-zA-Z0-9\-\_\.\/])+$`)

func validateMetadata(s string) bool {
	return validMetadata.MatchString(s)
}

func printfColored(col color.Attribute, str string) {
	_, err := color.New(col).Printf(str)
	if err != nil {
		common.Error(err)
	}
}

func authenticate(w http.ResponseWriter, r *http.Request) (*common.App, error) {
	if len(r.Header["X-Auth-Token"]) == 0 {
		http.Error(w, "", 401)
		printfColored(color.BgMagenta, "AUTHENTICATION ERROR")
		printfColored(color.FgMagenta, " X-Auth-Token header not found\n")
		return nil, errors.New("X-Auth-Token header not found")
	}
	header := r.Header["X-Auth-Token"][0]
	app, found := tokens[header]
	if !found {
		http.Error(w, "", 401)
		printfColored(color.BgMagenta, "AUTHENTICATION ERROR")
		printfColored(color.FgMagenta, " token did not match any known apps\n")
		return nil, errors.New("token did not match any known apps")
	}
	return &app, nil
}

func (wrapper fileWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app, err := authenticate(w, r)
	if err != nil {
		return
	}

	reader, status := wrapper.Handle(wrapper.Ctx, *app, r)
	// TODO: Zmiana tego statusu na prawilny
	if status != 200 {
		printfColored(color.BgRed, "ERROR: "+strconv.Itoa(status)+" "+r.RequestURI+"\n")
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=file.txt")
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	switch reader := reader.(type) {
	case io.ReadCloser:
		written, err := io.Copy(w, reader)
		if err != nil {
			common.Error(err)
		}
		logrus.Debug("Written ", written, " bytes.")
		err = reader.Close()
		if err != nil {
			common.Error(err)
		}
	}
}

func (wrapper jsonWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logrus.Debug("URI", r.RequestURI)
	app, err := authenticate(w, r)
	if err != nil {
		return
	}
	obj, status := wrapper.Handle(wrapper.Ctx, *app, r)
	str, err := json.Marshal(obj)
	if err != nil {
		http.Error(w, "could not encode json", 500)
		printfColored(color.BgMagenta, "JSON ERROR")
		printfColored(color.FgMagenta, " "+r.RequestURI+"\n")
		return
	}
	if status >= 400 {
		http.Error(w, string(str), status)
		printfColored(color.BgRed, "ERROR "+strconv.Itoa(status))
		printfColored(color.FgRed, " "+r.RequestURI+"\n")
		return
	}
	if status == 256 {
		//TODO zrobić to ładniej oraz dowiedzieć się co się dzieje po drodze
		w.Header().Set("Content-Disposition", "attachment; filename=file.txt")
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		switch rdr := obj.(type) {
		case io.ReadCloser:
			var written int64
			written, err = io.Copy(w, rdr)
			if err != nil {
				common.Error(err)
			}
			logrus.Debug("Written", written, "bytes.")
			err = rdr.Close()
			if err != nil {
				common.Error(err)
			}
		}
	}
	printfColored(color.BgGreen, "OK")
	printfColored(color.FgGreen, " "+r.RequestURI+"\n")
	_, err = w.Write(str)
	if err != nil {
		common.Error(err)
	}
}

func getPackageHandler(a *appContext, app common.App, r *http.Request) (interface{}, int) {
	/*
	 //TODO retrieve package info from mongo
	*/
	//id := r.URL.Query().Get("id")
	ret := make(map[string]interface{})
	ret["files"] = ""
	ret["metadata"] = ""
	return ret, 200
}

func postPackageHandler(a *appContext, app common.App, r *http.Request) (interface{}, int) {
	/*
		 action := r.Header.Get("X-Action")
		 if action == "copy" {
				 //TODO copy the package in mongo
				 return packageName, 200
		 }
	*/
	list, err := extractMetadata(r)
	if err != nil {
		return "could not extract metadata", 400
	}
	id, err := nextNumString(a)
	if err != nil {
		common.Error(err)
		return "", 500
	}
	pkg := common.Package{Prefix: configInstance.InstancePrefix, App: app.ID, ID: id}
	err = a.mongo.PutPackage(pkg, list)
	if err != nil {
		common.Error(err)
		return "", 500
	}
	fullQualifier := pkg.FullQualifier()
	logrus.Debug("FULLQ", fullQualifier)
	return fullQualifier, 200
}

func deletePackageHandler(a *appContext, app common.App, r *http.Request) (interface{}, int) {
	/*
	 //TODO delete package from mongo and files from storage
	*/
	return "", 200
}

func getFileHandler(a *appContext, app common.App, r *http.Request) (io.Reader, int) {
	id := r.URL.Query().Get("id")
	if id == "" {
		return nil, 400
	}
	var file common.File
	err := file.FromFullQualifier(id)
	if err != nil {
		return strings.NewReader(err.Error()), 400
	}
	reader, err := a.storage.GetFile(file, configInstance)
	if err != nil {
		return nil, 404
	}
	return reader, 200
}

func extractMetadata(r *http.Request) ([]common.Metadata, error) {
	metadata := make([]common.Metadata, 0)
	decoded, err := base64.StdEncoding.DecodeString(r.Header["X-Metadata"][0])
	if err != nil {
		return metadata, err
	}
	err = json.Unmarshal(decoded, &metadata)
	if err != nil {
		return metadata, err
	}
	for _, x := range metadata {
		if !validateMetadata(x.Key) || !validateMetadata(x.Value) {
			return metadata, errors.New("invalid metadata")
		}
	}
	return metadata, err
}

func postFileHandler(a *appContext, app common.App, r *http.Request) (interface{}, int) {
	reader, handler, err := r.FormFile("file")
	if err != nil {
		common.Error(err)
		return "something wrong with reader", 400
	}
	//TODO validate package and filename
	var pkg common.Package
	err = pkg.FromFullQualifier(r.Header["Package"][0])
	if err != nil {
		return err.Error(), 400
	}
	file := common.File{Package: pkg, ID: handler.Filename}

	metadata, err := extractMetadata(r)
	if err != nil {
		return "could not read metadata", 400
	}

	err = a.mongo.PutFile(file, metadata)
	if err != nil {
		return "could not put metadata", 500
	}

	//    if r.Header.Get("X-METADATA-TEXT") == "true" {
	//_ = common.StreamToString(reader)
	if r.Header.Get("Text") == "true" {
		var buf bytes.Buffer
		tee := io.TeeReader(reader, &buf)

		err = a.storage.PutFile(tee, file, configInstance)
		if err != nil {
			return "put file failed", 500
		}
		err = a.fulltext.AddText(file, &buf)
		if err != nil {
			return "fulltext add failed", 500
		}
	} else {
		err = a.storage.PutFile(reader, file, configInstance)
		if err != nil {
			common.Error(err)
			if strings.HasSuffix(err.Error(), "EOF") {
				return "EOF", 555
			}
			return "could not upload file", 500
		}
	}
	logrus.Debug("STWORZONO", file)

	return "ok", 200
}

func deleteFileHandler(a *appContext, app common.App, r *http.Request) (interface{}, int) {
	//TODO do we really want to pass the file for deletion by body?
	fullQualifier := common.StreamToString(r.Body)
	logrus.Debug("USUWAM", fullQualifier)
	var file common.File
	err := file.FromFullQualifier(fullQualifier)
	if err != nil {
		return strings.NewReader(err.Error()), 400
	}
	err = a.storage.DeleteFile(file, configInstance)
	if err != nil {
		if err.Error() == "Object Not Found" {
			return nil, 423 // locked
		}
		return nil, 500
	}
	err = a.fulltext.DeprecateDocument(fullQualifier)
	if err != nil {
		common.Error(err)
		return nil, 500
	}
	logrus.Debug("Do skasowania poszło:", file)
	return err, 200
}

func searchHandler(a *appContext, app common.App, r *http.Request) (interface{}, int) {
	//TODO search by mongo
	return "", 200
}

// Query is "boosted" by how aligned the words are. If the query is "a b" then every instance
// of "a b" rather than "b a" has higher score.
func fulltextPhraseHandler(a *appContext, app common.App, r *http.Request) (interface{}, int) {
	str := common.StreamToString(r.Body)
	res, err := a.fulltext.PhraseQuery(str)
	if err != nil {
		return err, 400
	}
	return res, 200
}

// Assumption - we get extracted (but not stemmed) document from clients. Result is the
// list of top N docs that satisfy search requirements. Results contain "beginning" position
// - that's where starting bow begins.
// No boost/tfidf in this query.
func fulltextDocHandler(a *appContext, app common.App, r *http.Request) (interface{}, int) {
	str := common.StreamToString(r.Body)
	res, err := a.fulltext.BowQuery(str)
	if err != nil {
		return err, 400
	}
	return res, 200
}

func fulltextRefreshHandler(a *appContext, app common.App, r *http.Request) (interface{}, int) {
	id := r.Header.Get("X-FileId")
	if id == "" {
		return nil, 400
	}
	var file common.File
	err := file.FromFullQualifier(id)
	if err != nil {
		return strings.NewReader(err.Error()), 400
	}
	reader, err := a.storage.GetFile(file, configInstance)
	if err != nil {
		common.Error(err)
		return nil, 404
	}

	err = a.fulltext.AddText(file, reader)
	if err != nil {
		common.Error(err)
		return nil, 500
	}

	return nil, 200
}
