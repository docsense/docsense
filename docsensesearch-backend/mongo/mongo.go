package mongo

import (
	"errors"
	sc_common "storage-core/common"
	"strconv"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var M *Mongo

func init() {
	m, err := NewMongo()
	M = &m
	if err != nil {
		sc_common.Error(err)
	}
}

//Mongo represents a mongodb connection
type Mongo struct {
	session *mgo.Session
}

//NewMongo creates a new Mongo instance
func NewMongo() (Mongo, error) {
	logrus.Info("creating new mongo")
	session, err := mgo.Dial("localhost")
	return Mongo{session}, err
}

//GetTuples gets tuples by primary key from the given collection
func (m *Mongo) GetTuples(selector bson.M, collection string) ([]bson.M, error) {
	session := m.session.Copy()
	defer session.Close()
	c := session.DB("docsensesearch").C(collection)
	q := c.Find(selector)
	result := make([]bson.M, 0)
	err := q.All(&result)
	return result, err
}

//GetTuple gets one tuple from the store, if there are more than 1 or there are none, it returns an error
func (m *Mongo) GetTuple(selector bson.M, collection string) (bson.M, error) {
	bsonMs, err := m.GetTuples(selector, collection)
	if err != nil {
		return nil, err
	}
	if len(bsonMs) != 1 {
		return nil, errors.New("wrong number of tuples returned: " + strconv.Itoa(len(bsonMs)))
	}
	return bsonMs[0], nil
}

//PutTuple puts a tuple in a mongo collection
func (m *Mongo) PutTuple(collection string, newtuple bson.M) (bson.ObjectId, error) {
	session := m.session.Copy()
	defer session.Close()
	c := session.DB("docsensesearch").C(collection)
	changeInfo, err := c.Upsert(newtuple, newtuple)
	if err != nil {
		return bson.ObjectId(""), err
	}
	if changeInfo.UpsertedId != nil {
		return changeInfo.UpsertedId.(bson.ObjectId), nil
	} else {
		return bson.ObjectId(""), nil
	}
}

//RemoveTuple (surprise, surprise!) removes a tuple.
func (m *Mongo) RemoveTuple(collection string, selector bson.M) error {
	bsonM, err := m.GetTuple(selector, collection)
	if err != nil {
		return err
	}
	session := m.session.Copy()
	c := session.DB("docsensesearch").C(collection)
	err = c.RemoveId(bsonM["_id"])
	return err
}
