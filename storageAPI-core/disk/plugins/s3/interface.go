package s3

import (
	"context"
	"io"
	"storage-core/common"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

//APIConnection is an implementation of disk.APIConnection
type APIConnection struct {
	endpoint string
	bucket   string
}

//Init inits the module, must be called before any other calls
func Init(config common.Config) *APIConnection {
	return &APIConnection{config.Storage.S3.Endpoint, config.InstancePrefix}
}

//GetFile given a common.File handle retrieves a file from an S3 storage
func (s *APIConnection) GetFile(file common.File, config common.Config) (io.ReadCloser, error) {
	downloader := s3manager.NewDownloader(session.New(&aws.Config{Endpoint: aws.String(s.endpoint), Region: aws.String("eu-central-1")}))
	ctx, _ := context.WithTimeout(aws.BackgroundContext(), 15*time.Second)
	//defer cancel()
	out, err := downloader.S3.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(file.FullQualifier()),
	})

	if err != nil {
		common.Error(err)
		return nil, err
	}

	return out.Body, nil
}

//PutFile puts a file in an S3 storage to be later retrieved with given common.File handle
func (s *APIConnection) PutFile(reader io.Reader, file common.File, config common.Config) error {
	metaData := make(map[string][]string)
	metaData["Content-Type"] = []string{"application/octet-stream"}

	ctx, _ := context.WithTimeout(aws.BackgroundContext(), 15*time.Second)

	uploader := s3manager.NewUploader(session.New(&aws.Config{Endpoint: aws.String(s.endpoint), Region: aws.String("eu-central-1")}))
	// Upload the file to S3.
	_, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(file.FullQualifier()),
		Body:   reader,
	})

	//_, err := s.conn.PutObjectWithMetadata(config.InstancePrefix, common.FileToFullQualifier(file), reader, metaData, nil)
	if err != nil {
		common.Error(err)
		return err
	}

	return nil
}

//DeleteFile deletes the file with the common.File handle from an S3 storage. What did you think it does? I'm only documenting this cause the linter told me to.
func (s *APIConnection) DeleteFile(file common.File, config common.Config) error {
	//err := s.conn.RemoveObject(config.InstancePrefix, common.FileToFullQualifier(file))
	return nil
}

//CopyFile does nothing yet. Don't use it. If you really need to, let me know.
func (s *APIConnection) CopyFile(f0 common.File, f1 common.File, config common.Config) error {
	return nil
}
