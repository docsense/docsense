package mongo

import (
	"errors"
	"storage-core/common"
	"strconv"

	"github.com/Sirupsen/logrus"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//Mongo represents a mongodb connection
type Mongo struct {
	session *mgo.Session
}

//NewMongo creates a new Mongo instance from config
func NewMongo(conf common.Config) (Mongo, error) {
	logrus.Info("creating new mongo")
	session, err := mgo.Dial("localhost")
	return Mongo{session}, err
}

func makeItem(mtd []common.Metadata) (map[string]interface{}, error) {
	item := make(map[string]interface{})
	for _, meta := range mtd {
		if meta.Type == "number" {
			n, err := strconv.Atoi(meta.Value)
			if err != nil {
				return nil, errors.New("could not convert string to int")
			}
			item[meta.Key] = n
		} else if meta.Type == "string" {
			item[meta.Key] = meta.Value
		} else {
			return nil, errors.New("unknown type " + meta.Type)
		}
	}
	return item, nil
}

//PutPackage puts the common.Package with the given metadata in the mongo instance
func (mongo *Mongo) PutPackage(pkg common.Package, mtd []common.Metadata) error {
	session := mongo.session.Copy()
	defer session.Close()
	c := session.DB("store").C(pkg.Prefix + "." + pkg.App)
	item, err := makeItem(mtd)
	if err != nil {
		return err
	}
	item["_id"] = pkg.FullQualifier()
	err = c.Insert(item)
	return err
}

//PutFile puts the common.File with the given metadata in the mongo instance
func (mongo *Mongo) PutFile(file common.File, mtd []common.Metadata) error {
	session := mongo.session.Copy()
	defer session.Close()
	c := session.DB("store").C(file.Package.Prefix + "." + file.Package.App)
	item, err := makeItem(mtd)
	if err != nil {
		return err
	}
	item["_id"] = file.FullQualifier()
	err = c.Insert(item)
	return err
}

//GetTuples gets tuples by primary key from the primary store
func (mongo *Mongo) GetTuples(key string) ([]bson.M, error) {
	session := mongo.session.Copy()
	defer session.Close()
	c := session.DB("store").C("storagecore")
	q := c.FindId(key)
	result := make([]bson.M, 0)
	err := q.All(&result)
	if len(result) == 0 {
		return nil, errors.New("Empty shit")
	}
	return result, err
}

//PutTuple puts a tuple in the primary store by key
func (mongo *Mongo) PutTuple(key string, update bson.M) error {
	session := mongo.session.Copy()
	defer session.Close()
	c := session.DB("store").C("storagecore")
	_, err := c.UpsertId(key, update)
	return err
}
