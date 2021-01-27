package storage

type Storage interface {
	Save(fileName, contentType string, buf *[]byte) (string, error)
}
