package controllers

import (
	"docsensesearch/mongo"
	"docsensesearch/storagecore"
	"os/exec"
	sc_common "storage-core/common"
	"strconv"

	"gopkg.in/mgo.v2/bson"

	"github.com/Sirupsen/logrus"
	"github.com/astaxie/beego"
)

//MainController just tells you that you can't access root
type MainController struct {
	beego.Controller
}

//RefreshSearchController is an internal tool to fill fulltext from ceph
type RefreshSearchController struct {
	beego.Controller
}

//ManageSpListsController operates adding sharepoint lists for the crawler
type ManageSpListsController struct {
	beego.Controller
}

//Get just tells you that you can't access root
func (main *MainController) Get() {
	main.Data["json"] = map[string]interface{}{"status": "error", "error": "tried to access root"}
	main.ServeJSON()
}

//Post operates adding sharepoint lists for the crawler
func (listManager *ManageSpListsController) Post() {
	spLink := listManager.GetString("sp_link")
	spID := listManager.GetString("sp_id")
	logrus.Debug(spLink, spID)
	//TODO this is probably a serious security hazard
	out, err := exec.Command("/usr/bin/python", "utilscripts/test_sp.py", spLink, spID).Output()
	if err != nil {
		sc_common.Error(err)
		listManager.Ctx.Output.Status = 500
		listManager.ServeJSON()
		return
	}
	num, err := strconv.Atoi(string(out)[:len(string(out))-1])
	if err != nil {
		sc_common.Error(err)
		listManager.Ctx.Output.Status = 500
		listManager.ServeJSON()
		return
	}
	listManager.Data["json"] = map[string]interface{}{"correct_entries": num}
	if num >= 0 {
		//ignore return value of _id
		_, err := mongo.M.PutTuple("sp_lists", bson.M{
			"id":             0,
			"link":           spLink,
			"sp_id":          spID,
			"last_migration": "1970-01-01T01:00:00",
		})
		if err != nil {
			sc_common.Error(err)
			listManager.Ctx.Output.Status = 500
			listManager.ServeJSON()
			return
		}
	}
	listManager.ServeJSON()
}

//SpList is the format of data passed to client
type SpList struct {
	ID            int
	Link          string
	SpID          string
	LastMigration string
}

/*
func paramsListsToSpLists(paramsLists []orm.ParamsList) ([]SpList, error) {
	ret := make([]SpList, 0)
	//layout := "1970-01-01 01:00:00"
	layout := "2006-01-02 15:04:05"
	for _, row := range paramsLists {
		id, err0 := strconv.Atoi(row[0].(string))
		time, err1 := time.Parse(layout, row[3].(string))

		if err0 != nil || err1 != nil {
			sc_common.Error(err0, err1)
			return nil, errors.New("conversion error")
		}
		ret = append(ret, SpList{
			id,              // id
			row[1].(string), // filename
			row[2].(string), // sc_id
			time,            // link
		})
	}
	return ret, nil
}*/

func bsonMsToSpLists(bsons []bson.M) ([]SpList, error) {
	ret := make([]SpList, 0)
	for _, bsonM := range bsons {
		ret = append(ret, SpList{
			bsonM["id"].(int),
			bsonM["link"].(string),           // link
			bsonM["sp_id"].(string),          // sp_id
			bsonM["last_migration"].(string), // last_migration
		})
	}
	return ret, nil
}

//Get returns all sp lists currently in the database
func (listManager *ManageSpListsController) Get() {
	bsonMs, err := mongo.M.GetTuples(bson.M{}, "sp_lists")
	if err != nil {
		sc_common.Error(err)
		listManager.Ctx.Output.Status = 500
		listManager.ServeJSON()
		return
	}

	spLists, err := bsonMsToSpLists(bsonMs)

	if err != nil {
		sc_common.Error(err)
		listManager.Ctx.Output.Status = 500
		listManager.ServeJSON()
		return
	}

	listManager.Data["json"] = spLists
	listManager.ServeJSON()
	//linter hacks
	//fake use SpList struct fields
	//we need these fields to pass them to frontend, but they are used implicitly and it is not recognized by the linter
	//ofc this will be optimized out by the compiler
	var fakeStruct SpList
	_ = fakeStruct.ID
	_ = fakeStruct.Link
	_ = fakeStruct.SpID
	_ = fakeStruct.LastMigration
}

//Get operates the refresh, to be curled
func (refresher *RefreshSearchController) Get() {
	files := make([]string, 0)
	bsonMs, err := mongo.M.GetTuples(nil, "files")
	if err != nil {
		sc_common.Error(err)
		refresher.Ctx.Output.Status = 500
		refresher.ServeJSON()
		return
	}
	for _, bsonM := range bsonMs {
		files = append(files, bsonM["sc_id"].(string))
	}
	for _, scID := range files {
		logrus.Debug("refreshing", scID)
		hds := []storagecore.Header{
			{
				Key:   "X-FileId",
				Value: scID + ".text",
			},
		}
		var err error
		_, err = storagecore.Request(nil, "POST", "/fulltext/refresh", hds)
		if err != nil {
			sc_common.Error(err)
			refresher.Ctx.Output.Status = 500
			refresher.ServeJSON()
			return
		}
	}
	refresher.ServeJSON()
	return
}
