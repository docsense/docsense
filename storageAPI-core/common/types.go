package common

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"runtime"

	"github.com/Sirupsen/logrus"
)

//Config is the format of the config.json file, the general configuration file for the app
type Config struct {
	Apps []App
	HTTP struct {
		Addr string
		Port string
	}
	Storage struct {
		S3 struct {
			Endpoint        string
			AccessKeyID     string
			SecretAccessKey string
			UseSSL          bool
		}
	}
	Fulltext struct {
		Addr string
		Port string
	}
	InstancePrefix string
}

//App represents an application. Id is a short name (e.g. "docsense"), Token is a randomly generated security token
type App struct {
	ID    string
	Token string
}

func (a *App) String() string {
	return a.ID
}

//Metadata represents a single instance of a metadata. Type can be "number" or "string", it alters the behaviour of Mongo searches
type Metadata struct {
	Key   string
	Value string
	Type  string
}

//Package represents a storage core package. Prefix is the instance prefix, App is the app the package is assiociated with
type Package struct {
	Prefix string
	App    string
	ID     string
}

//File unequivocally represents a single file within a package. It is equivalent to a storage (Ceph/S3) object
type File struct {
	Package Package
	ID      string
}

//FromFullQualifier reads a string full qualifier ("instance.application.id.filename") to the common.File
func (file *File) FromFullQualifier(str string) error {
	l := strings.Split(str, ".")
	if len(l) < 4 {
		return errors.New("wrong format of file full qualifier " + str)
	}
	file.Package = Package{l[0], l[1], l[2]}
	file.ID = l[3]
	return nil
}

//FromFullQualifier reads a string full qualifier ("instance.application.id") to the common.Package
func (pkg *Package) FromFullQualifier(str string) error {
	l := strings.Split(str, ".")
	if len(l) < 3 {
		return errors.New("wrong format of package full qualifier " + str)
	}
	pkg.Prefix = l[0]
	pkg.App = l[1]
	pkg.ID = l[2]
	return nil
}

//FullQualifier returns the full qualifier ("instance.application.id.filename")
func (file *File) FullQualifier() string {
	return file.Package.FullQualifier() + "." + file.ID
}

//FullQualifier returns the full qualifier ("instance.application.id")
func (pkg *Package) FullQualifier() string {
	return pkg.Prefix + "." + pkg.App + "." + pkg.ID
}

func (m *Metadata) String() string {
	return m.Key + ": " + m.Value
}

//StreamToString fully reads a reader to a string
func StreamToString(stream io.Reader) string {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(stream)
	if err != nil {
		Error(err)
	}
	return buf.String()
}

//Error logs errors in a unified way using logrus
func Error(errs ...error) {
	pc, file, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	errStrs := make([]string, 0)
	for _, err := range errs {
		errStrs = append(errStrs, err.Error())
	}
	logrus.WithFields(logrus.Fields{
		"err":  strings.Join(errStrs, ", "),
		"line": line,
		"func": f.Name(),
		"file": file,
	}).Error()
}
