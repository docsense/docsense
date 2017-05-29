package controllers

import (
	"docsensesearch/mongo"
	"docsensesearch/storagecore"
	"encoding/base64"
	sc_common "storage-core/common"
	"strings"

	"gopkg.in/mgo.v2/bson"

	"fmt"

	"errors"

	"github.com/astaxie/beego"
)

//DownloadController serves PDFs to frontend
type DownloadController struct {
	beego.Controller
}

func (downloader *DownloadController) servePDF(data []byte, filename string) {
	downloader.Ctx.Output.Header("Content-Description", "File Transfer")
	downloader.Ctx.Output.Header("Content-Type", "application/pdf")
	downloader.Ctx.Output.Header("Content-Disposition", "attachment; filename="+filename)
	downloader.Ctx.Output.Header("Content-Transfer-Encoding", "binary")
	downloader.Ctx.Output.Header("Expires", "0")
	downloader.Ctx.Output.Header("Cache-Control", "must-revalidate")
	downloader.Ctx.Output.Header("Pragma", "public")
	err := downloader.Ctx.Output.Body(data)
	if err != nil {
		sc_common.Error(err)
	}
}

//Get serves a PDF to frontend
func (downloader *DownloadController) Get() {
	fileid := downloader.GetString("id")
	fmt.Println(fileid)
	objIDBytes, err := base64.StdEncoding.DecodeString(fileid)
	if err != nil {
		sc_common.Error(err)
		downloader.Ctx.Output.Status = 500
		downloader.ServeJSON()
		return
	}
	bsonMs, err := mongo.M.GetTuples(bson.M{"_id": bson.ObjectId(objIDBytes)}, "files")
	if len(bsonMs) != 1 {
		sc_common.Error(errors.New("wrong number of entries found"))
		downloader.Ctx.Output.Status = 500
		downloader.ServeJSON()
		return
	}
	myFile := bsonMs[0]
	data, err := storagecore.GetFile(myFile["sc_id"].(string), "file")
	if err != nil {
		sc_common.Error(err)
		downloader.Ctx.Output.Status = 500
		downloader.ServeJSON()
		return
	}
	l := strings.Split(myFile["link"].(string), "/")
	filename := l[len(l)-1]
	downloader.servePDF(data, filename)
	return
}
