package disk

import (
	"io"
	"storage-core/common"
)

//APIConnection is an abstraction of a storage device, it can be an S3 API, a Swift API, files stored on disk etc.
type APIConnection interface {
	GetFile(common.File, common.Config) (io.ReadCloser, error)
	CopyFile(common.File, common.File, common.Config) error
	PutFile(io.Reader, common.File, common.Config) error
	DeleteFile(common.File, common.Config) error
}
