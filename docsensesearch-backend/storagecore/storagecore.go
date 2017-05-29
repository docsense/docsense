package storagecore

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	sc_common "storage-core/common"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/astaxie/beego"
)

//Header defines the Header format for storagecore
type Header struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
	Type  string `json:"Type"`
}

var netClient = &http.Client{
	Timeout: time.Second * 20,
	Transport: &http.Transport{
		DisableKeepAlives:   true,
		MaxIdleConnsPerHost: 100,
	},
}

//PostFile sends a file to storagecore with given name, package, textness and contents
//TODO add support for file metadata
func PostFile(name string, pkg string, text bool, rd io.Reader) error {
	body, contenttype, err := prepareMultipart(name, rd)
	if err != nil {
		return err
	}

	metadata := []sc_common.Metadata{
		{
			Key:   "metadan",
			Value: "wartość metadana",
			Type:  "string",
		},
		{
			Key:   "inny metadan",
			Value: "2137",
			Type:  "number",
		},
	}

	byteMetadata, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	encoded := base64.StdEncoding.EncodeToString(byteMetadata)

	hds := []Header{
		{
			Key:   "Package",
			Value: pkg,
		},
		{
			Key:   "Content-Type",
			Value: contenttype,
		},
		{
			Key:   "X-METADATA",
			Value: encoded,
		},
	}

	if text {
		hds = append(hds, Header{Key: "text", Value: "true"})
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(body)
	if err != nil {
		return err
	}

	_, err = Request(bytes.NewReader(buf.Bytes()), "POST", "/file", hds)
	if err != nil {
		sc_common.Error(err)
	}

	return nil
}

//DeleteDocument does... take your best guess. Fuck you, linter.
func DeleteDocument(scID string) error {
	hds := []Header{
		{
			Key:   "X-DELETE-CASCADE",
			Value: "true",
		},
	}
	_, err := Request(strings.NewReader("\""+scID+"\""), "DELETE", "/package", hds)
	return err
}

func prepareMultipart(name string, rd io.Reader) (io.Reader, string, error) {
	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		return nil, "", err
	}
	_, err = io.Copy(part, rd)
	if err != nil {
		return nil, "", err
	}
	err = writer.Close()
	if err != nil {
		return nil, "", err
	}

	return &body, writer.FormDataContentType(), nil
}

//Request makes a general request to storage-core
func Request(body io.Reader, method, endpoint string, headers []Header) ([]byte, error) {
	req, err := http.NewRequest(method, beego.AppConfig.String("storagecoreendpoint")+endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-AUTH-TOKEN", beego.AppConfig.String("storagecoretoken"))
	for _, x := range headers {
		req.Header.Set(x.Key, x.Value)
	}
	response, err := netClient.Do(req)
	if err != nil {
		return nil, err
	} else if response.StatusCode != 200 {
		return nil, errors.New("Wrong status code: " + strconv.Itoa(response.StatusCode))
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(response.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

//PostPackage creates a new package with the given metadata and returns the created index.
func PostPackage(metadata []sc_common.Metadata) (string, error) {
	metadataBytes, err := json.Marshal(metadata)
	encoded := base64.StdEncoding.EncodeToString(metadataBytes)
	hds := []Header{
		{
			Key:   "X-METADATA",
			Value: encoded,
		},
	}
	if err != nil {
		return "", err
	}
	bts, err := Request(nil, "POST", "/package", hds)
	if err != nil {
		return "", err
	}
	return string(bts), nil
}

//GetFile retrieves a file from the given bucket (in the form of full qualifier) with the given filename
func GetFile(bucket string, filename string) ([]byte, error) {
	logrus.Debug("getfile", bucket, filename)
	fileQualifier := bucket + "." + filename
	bts, err := Request(nil, "GET", "/file?id="+fileQualifier, nil)
	if err != nil {
		return nil, err
	}
	return bts, nil
}
