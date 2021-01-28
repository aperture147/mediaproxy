package storage

import (
	"io/ioutil"
	"path"
)

type FileSystemStorage struct {
	Path string
}

func NewFileSystemStorage(path string) Storage {
	return FileSystemStorage{
		Path: path,
	}
}

/*
fileName here should be a full path with extension of the file
this function returns full path of the uploaded file
saving file to
*/
func (s FileSystemStorage) Save(fileName, _ string, buf *[]byte) (string, error) {
	fullPath := path.Join(s.Path, fileName)
	err := ioutil.WriteFile(fullPath, *buf, 0755)
	return fullPath, err
}
