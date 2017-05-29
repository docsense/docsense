package upload

import (
	"bytes"
	"docsensesearch/mongo"
	"docsensesearch/storagecore"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	sc_common "storage-core/common"

	"errors"

	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/astaxie/beego"
)

var netClient = &http.Client{
	Timeout: time.Second * 60 * 10,
	Transport: &http.Transport{
		DisableKeepAlives:   true,
		MaxIdleConnsPerHost: 100,
	},
}

//Controller handles uploading files to the system
type Controller struct {
	beego.Controller
}

func createPackageForUserFile(filename string) (string, error) {
	pkgID, err := storagecore.PostPackage([]sc_common.Metadata{{"filename", filename, "string"},
		{"othermetadata", strconv.Itoa(time.Now().Second()), "number"}})
	if err != nil {
		return "", err
	}
	pkgID = pkgID[1 : len(pkgID)-1]
	return pkgID, nil
}

//RunConvertRequest runs a request (synchronically) to docd to convert a pdf to text and returns the result
func RunConvertRequest(file io.Reader) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("input", "pdf.pdf")
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", "http://localhost:5008/convert", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := netClient.Do(req)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}
	var convertResponse map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &convertResponse)
	if err != nil {
		sc_common.Error(err)
		return "", err
	}
	return convertResponse["body"].(string), nil
}

type readSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

func runUploadFileRequest(filename string, file readSeekCloser) (string, error) {
	pkgID, err := createPackageForUserFile(filename)
	if err != nil {
		return "", err
	}
	err = storagecore.PostFile("file", pkgID, false, file)
	if err != nil {
		return "", err
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	cp := io.TeeReader(file, &buf)

	text, err := RunConvertRequest(cp)
	if err != nil {
		return "", err
	}

	err = file.Close()
	if err != nil {
		sc_common.Error(err)
	}
	textReader := strings.NewReader(text)

	err = storagecore.PostFile("text", pkgID, true, textReader)

	return pkgID, err
}

var oneLetter = regexp.MustCompile(`[a-zA-Z0-9 żółćęśąźńŻÓŁĆĘŚĄŹŃ\.]`)

func cleanFilename(filename string) string {
	cleanFilename := strings.Map(func(r rune) rune {
		b := oneLetter.MatchString(string(r))
		if b {
			return r
		}
		return -1
	}, filename)
	return cleanFilename
}

//Post handles uploading files by a POST request
func (controller *Controller) Post() {
	r := controller.Ctx.Request
	dllink := controller.GetString("dllink")
	filename := controller.GetString("filename")
	spID := controller.GetString("spid")
	spListID := controller.GetString("splistid")
	year := controller.GetString("year")
	position := controller.GetString("position")
	date := controller.GetString("date")
	isNew := controller.GetString("isnew")
	date = date[:len(date)-1] // ugly hack FIXME
	if dllink == "" {
		sc_common.Error(errors.New("no download link"))
		controller.Ctx.Output.Status = 400
		controller.ServeJSON()
		return
	}
	if isNew != "True" {
		//the file is already in the database, it has been modified
		// TODO implement handling of this case
		oldFile, err := mongo.M.GetTuple(bson.M{"sp_id": spID, "sp_list": spListID}, "files")
		if err != nil {
			sc_common.Error(err)
			controller.Ctx.Output.Status = 500
			controller.ServeJSON()
			return
		}
		err = storagecore.DeleteDocument(oldFile["sc_id"].(string))
		if err != nil {
			sc_common.Error(err)
			controller.Ctx.Output.Status = 500
			controller.ServeJSON()
			return
		}
		err = mongo.M.RemoveTuple("files", bson.M{"sp_id": spID, "sp_list": spListID})
		if err != nil {
			sc_common.Error(err)
			controller.Ctx.Output.Status = 500
			controller.ServeJSON()
		}
		return
	}

	fh := r.MultipartForm.File["file"][0]
	f, err := fh.Open()
	fmt.Println(f)
	if err != nil {
		sc_common.Error(err)
		controller.Ctx.Output.Status = 500
		controller.ServeJSON()
		return
	}

	filename = cleanFilename(filename)
	pkgID, err := runUploadFileRequest(filename, f)
	if err != nil {
		sc_common.Error(err)
		controller.Ctx.Output.Status = 500
		controller.ServeJSON()
		//return
	}
	logrus.Debug("FILENAME", filename)
	_, err = mongo.M.PutTuple("files", bson.M{
		"filename": filename,
		"sc_id":    pkgID,
		"link":     dllink,
		"sp_id":    spID,
		"sp_list":  spListID,
		"position": position,
		"year":     year,
		"date":     date,
	})
	if err != nil {
		sc_common.Error(err)
		controller.Ctx.Output.Status = 500
		controller.ServeJSON()
		return
	}

	controller.Data["json"] = nil
	controller.ServeJSON()
}
