package search

import (
	"bytes"
	"docsensesearch/mongo"
	"docsensesearch/storagecore"
	"docsensesearch/upload"
	"encoding/base64"
	"encoding/json"
	"errors"
	sc_common "storage-core/common"
	"strconv"
	"strings"

	"gopkg.in/mgo.v2/bson"

	"fmt"

	"github.com/astaxie/beego"
)

//Controller handles searches by queries
type Controller struct {
	beego.Controller
}

//ByFileController unsurprisingly, handles searches by files
type ByFileController struct {
	beego.Controller
}

type fulltextResponse struct {
	Result []struct {
		Doc    string `json:"Doc"`
		Score  int    `json:"Score"`
		PosBeg int    `json:"PosBeg"`
		PosEnd int    `json:"PosEnd"`
	} `json:"Result"`
}

type ssFileWithLink struct {
	ID       string
	Filename string
	ScID     string
	Link     string
	SpID     int
	SpList   int
	Position int
	Year     int
	Date     string
	Koala    string
}

func min(x int, y int) int {
	if x < y {
		return x
	}
	return y
}

func max(x int, y int) int {
	if x > y {
		return x
	}
	return y
}

var wordFuzz = 5

func getWordNeighbourhoods(filePositions map[string][]int) (map[string][]string, error) {
	res := make(map[string][]string)
	for fileName, positions := range filePositions {
		if fileName == "" {
			return nil, errors.New("neighbourhoods file scid empty")
		}
		bytez, err := storagecore.GetFile(fileName, "text")
		if err != nil {
			return nil, err
		}
		words := strings.Fields(string(bytez))
		excerpts := make([]string, 0)
		for _, pos := range positions {
			if len(words) < pos {
				return res, errors.New("pos out of bounds" + strconv.Itoa(pos) + " " + fileName + " " + strings.Join(words, " "))
			}
			excerpt := strings.Join(words[max(pos-wordFuzz, 0):min(pos+wordFuzz, len(words)-1)], " ")
			excerpts = append(excerpts, excerpt)
		}
		res[fileName] = excerpts
	}
	return res, nil
}

func bsonMsToSSFilesWithLinks(bsons []bson.M) (map[string]ssFileWithLink, error) {
	ret := make(map[string]ssFileWithLink)
	for _, bsonM := range bsons {
		spID, err1 := strconv.Atoi(bsonM["sp_id"].(string))
		spList, err2 := strconv.Atoi(bsonM["sp_list"].(string))
		position, err3 := strconv.Atoi(bsonM["position"].(string))
		year, err4 := strconv.Atoi(bsonM["year"].(string))
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			sc_common.Error(err1, err2, err3, err4)
			return nil, errors.New("conversion error")
		}
		koalaBson, err := mongo.M.GetTuples(bson.M{
			"id": spList,
		}, "sp_lists")
		if err != nil {
			return nil, err
		}
		if len(koalaBson) != 1 {
			fmt.Println(bsonM)
			return nil, errors.New("sp_list not found for sp_id " + strconv.Itoa(spList))
		}
		objectID := bsonM["_id"].(bson.ObjectId)
		base64dObjectID := base64.StdEncoding.EncodeToString([]byte(objectID))
		ret[bsonM["sc_id"].(string)] = ssFileWithLink{
			base64dObjectID,            // ObjectId
			bsonM["filename"].(string), // filename
			bsonM["sc_id"].(string),    // sc_id
			bsonM["link"].(string),     // link
			spID,     // sp_id
			spList,   // sp_list
			position, // position
			year,     // year
			bsonM["date"].(string),        // date - this is not a string?...
			koalaBson[0]["link"].(string), // koala
		}
	}
	return ret, nil
}

type searchResult struct {
	ssFileWithLink
	Texts []string
	Score struct{ S, Beg, End int }
}

func processSearchResponse(ftr fulltextResponse) ([]searchResult, error) {
	res := make([]searchResult, 0)
	if len(ftr.Result) == 0 {
		return res, nil
	}

	scores := make(map[string]struct{ S, Beg, End int })

	fileNames := make([]string, 0)
	for _, d := range ftr.Result {
		fileName := strings.Join(strings.Split(d.Doc, ".")[0:3], ".")
		fileNames = append(fileNames, fileName)
		scores[fileName] = struct{ S, Beg, End int }{d.Score, d.PosBeg, d.PosEnd}
	}

	orList := make([]bson.M, 0)
	for _, fileName := range fileNames {
		orList = append(orList, bson.M{
			"sc_id": fileName,
		})
	}

	query := bson.M{"$or": orList}

	bsonMs, err := mongo.M.GetTuples(query, "files")

	if err != nil {
		sc_common.Error(err)
		return nil, err
	}
	files, err := bsonMsToSSFilesWithLinks(bsonMs)
	if err != nil {
		sc_common.Error(err)
		return nil, err
	}
	filePositions := make(map[string][]int)

	for _, result := range ftr.Result {
		fileName := strings.Join(strings.Split(result.Doc, ".")[0:3], ".")
		list := filePositions[fileName]
		list = append(list, result.PosBeg)
		if result.PosBeg != result.PosEnd {
			list = append(list, result.PosEnd)
		}
		filePositions[fileName] = list
	}

	fileTexts, err := getWordNeighbourhoods(filePositions)
	if err != nil {
		return nil, err
	}

	for fileName, texts := range fileTexts {
		res = append(res, searchResult{
			files[fileName],
			texts,
			scores[fileName],
		})
	}

	return res, nil
}

//Post handles searches by phrase
func (search *Controller) Post() {
	phrase := search.GetString("phrase")

	response, err := storagecore.Request(bytes.NewReader([]byte(phrase)), "POST", "/fulltext/phrase", nil)
	if err != nil {
		sc_common.Error(err)
		search.Ctx.Output.Status = 500
		search.ServeJSON()
		return
	}

	var ftr fulltextResponse
	err = json.Unmarshal(response, &ftr)
	if err != nil {
		sc_common.Error(err)
		search.Ctx.Output.Status = 500
		search.ServeJSON()
		return
	}

	res, err := processSearchResponse(ftr)
	if err != nil {
		sc_common.Error(err)
		search.Ctx.Output.Status = 500
		search.ServeJSON()
		return
	}

	search.Data["json"] = res
	search.ServeJSON()
}

func processFileSearchResponse(ftr fulltextResponse) ([]interface{}, error) {
	bsonMs, err := mongo.M.GetTuples(bson.M{}, "files")

	res := make([]interface{}, 0)

	myFiles := make(map[string]ssFileWithLink)

	files, err := bsonMsToSSFilesWithLinks(bsonMs)

	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.ScID == "" {
			return nil, errors.New("search response scid empty")
		}
		myFiles[file.ScID] = file
	}

	for _, result := range ftr.Result {
		pkg := strings.Join(strings.Split(result.Doc, ".")[0:3], ".")
		file := myFiles[pkg]
		res = append(res, struct {
			ssFileWithLink
			Texts []string
		}{
			file,
			[]string{},
		})
	}
	return res, nil
}

//Post handles searches by file
func (searchByFile *ByFileController) Post() {
	r := searchByFile.Ctx.Request
	if len(r.MultipartForm.File["file"]) == 0 {
		sc_common.Error(errors.New("no file uploaded"))
		searchByFile.Ctx.Output.Status = 400
		searchByFile.ServeJSON()
		return
	}
	fh := r.MultipartForm.File["file"][0]
	file, err := fh.Open()
	if err != nil {
		sc_common.Error(err)
		searchByFile.Ctx.Output.Status = 500
		searchByFile.ServeJSON()
		return
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		sc_common.Error(err)
		searchByFile.Ctx.Output.Status = 500
		searchByFile.ServeJSON()
		return
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(file)
	if err != nil {
		sc_common.Error(err)
		searchByFile.Ctx.Output.Status = 500
		searchByFile.ServeJSON()
		return
	}
	err = file.Close()
	if err != nil {
		sc_common.Error(err)
		searchByFile.Ctx.Output.Status = 500
		searchByFile.ServeJSON()
		return
	}
	text, err := upload.RunConvertRequest(bytes.NewReader(buf.Bytes()))
	if err != nil {
		sc_common.Error(err)
		searchByFile.Ctx.Output.Status = 500
		searchByFile.ServeJSON()
		return
	}

	response, err := storagecore.Request(bytes.NewReader([]byte(text)), "POST", "/fulltext/bow", nil)
	if err != nil {
		sc_common.Error(err)
		searchByFile.Ctx.Output.Status = 500
		searchByFile.ServeJSON()
		return
	}

	var ftr fulltextResponse
	err = json.Unmarshal(response, &ftr)
	if err != nil {
		sc_common.Error(err)
		searchByFile.Ctx.Output.Status = 500
		searchByFile.ServeJSON()
		return
	}

	res, err := processFileSearchResponse(ftr)
	if err != nil {
		searchByFile.Ctx.Output.Status = 500
		searchByFile.ServeJSON()
		return
	}

	searchByFile.Data["json"] = res
	searchByFile.ServeJSON()
}
