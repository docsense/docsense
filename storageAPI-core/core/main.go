package main

import (
	"net/http"
	"storage-core/common"
	"storage-core/mongo"
	"strconv"

	"storage-core/disk/plugins/s3"
	"storage-core/search/fulltext"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	"errors"

	"github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
)

func getConfig() common.Config {
	var config common.Config
	viper.SetConfigType("json")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()

	if err != nil {
		common.Error(err)
		return common.Config{}
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		common.Error(err)
		return common.Config{}
	}
	return config
}

func notFound(w http.ResponseWriter, r *http.Request) {
	printfColored(color.BgMagenta, "ERROR 404 NOT FOUND")
	printfColored(color.FgMagenta, " "+r.RequestURI+"\n")
	http.Error(w, "404 not found milordzie", 404)
}

var configInstance common.Config

func setMaxFile(m *mongo.Mongo) error {
	items, err := m.GetTuples("maxdocument")
	if err != nil {
		common.Error(err)
		return err
	}
	if len(items) != 1 {
		return errors.New("wrong count of items maxdocument " + strconv.Itoa(len(items)))
	}
	maxFile = items[0]["maxdocument"].(int) // if a document by _id maxdocument exists I assume it is correct
	return nil
}

func main() {
	logrus.SetLevel(logrus.DebugLevel) // you probably want to remove this to make the output more readable

	configInstance = getConfig()

	if len(configInstance.Apps) == 0 {
		logrus.WithFields(logrus.Fields{
			"func": "main",
			"err":  "no apps loaded",
		}).Fatal()
	}

	initTokens(configInstance)

	storage := s3.Init(configInstance)

	ft := fulltext.Create(configInstance)

	m, err := mongo.NewMongo(configInstance)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"func": "main",
			"err":  err,
		}).Fatal()
	}

	err = setMaxFile(&m)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"func": "setMaxFile",
			"err":  err,
		}).Fatal()
	}

	context := &appContext{
		storage:  storage,
		fulltext: &ft,
		mongo:    &m,
	}

	http.DefaultClient.Timeout = 10 * 60 * time.Second

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(notFound)
	r.Handle("/package", jsonWrapper{context, getPackageHandler}).Methods("GET")
	r.Handle("/package", jsonWrapper{context, postPackageHandler}).Methods("POST")
	r.Handle("/package", jsonWrapper{context, deletePackageHandler}).Methods("DELETE")
	r.Handle("/file", fileWrapper{context, getFileHandler}).Methods("GET")
	r.Handle("/file", jsonWrapper{context, postFileHandler}).Methods("POST")
	r.Handle("/file", jsonWrapper{context, deleteFileHandler}).Methods("DELETE")
	r.Handle("/search", jsonWrapper{context, searchHandler}).Methods("POST")
	r.Handle("/fulltext/phrase", jsonWrapper{context, fulltextPhraseHandler}).Methods("POST")
	r.Handle("/fulltext/bow", jsonWrapper{context, fulltextDocHandler}).Methods("POST")
	r.Handle("/fulltext/refresh", jsonWrapper{context, fulltextRefreshHandler}).Methods("POST")

	spew.Dump(configInstance)
	logrus.Info("starting application!")

	s := &http.Server{
		Addr:           configInstance.HTTP.Addr + ":" + configInstance.HTTP.Port,
		Handler:        r,
		ReadTimeout:    10 * 60 * time.Second,
		WriteTimeout:   10 * 60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err = s.ListenAndServe()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"func": "main",
			"err":  err,
		}).Fatal()
	}
}
