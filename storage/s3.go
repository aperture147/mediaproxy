package storage

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"path"
)

var AllUserAccess = aws.String("uri=http://acs.amazonaws.com/groups/global/AllUsers")

type S3Storage struct {
	Uploader *s3manager.Uploader // thread-safe uploader
	Bucket   *string             // selected bucket
	Path     *string             // path inside the bucket
}

func NewS3Storage(bucket, path string) Storage {
	sess := session.Must(session.NewSession())
	return S3Storage{
		Uploader: s3manager.NewUploader(sess),
		Bucket:   &bucket,
		Path:     &path,
	}
}

/*
fileName here should be a full path with extension of the file
contentType must be set to ensure the file type
this function returns full path of uploaded file, included the bucket path
*/
func (s S3Storage) Save(fileName, contentType string, buf *[]byte) (string, error) {
	key := aws.String(path.Join(*s.Path, fileName))
	_, err := s.Uploader.Upload(&s3manager.UploadInput{
		Bucket:      s.Bucket,
		Key:         key,           // Actually file path in bucket
		GrantRead:   AllUserAccess, // everyone can read this file
		Body:        bytes.NewReader(*buf),
		ContentType: aws.String(contentType),
	})
	return *key, err
}
