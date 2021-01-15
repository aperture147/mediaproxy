package middleware

import (
	"context"
	"errors"
	"io/ioutil"
	"mediaproxy/util"
	"net/http"
)

// Check the size of a request then check the size of a defined field of that request
// and finally put the selected field to the request context
type FileExtractor struct {
	AllowedSize int64  // data size in byte
	Field       string // the field which needs to be checked
}

type Sizer interface {
	Size() int64
}

func NewFileExtractor(size int, field string) FileExtractor {
	return FileExtractor{int64(size * 1024 * 1024), field}
}

// This function will try to verity the whole package size
func (fe FileExtractor) Verify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(fe.AllowedSize); err != nil {
			util.WriteBadRequestResponse(w, err)
			return
		}

		file, _, err := r.FormFile(fe.Field)

		if err != nil {
			util.WriteBadRequestResponse(w, err)
			return
		}

		if file.(Sizer).Size() > fe.AllowedSize {
			util.WriteBadRequestResponse(w, errors.New(fe.Field+" too large"))
			return
		}

		buffer, err := ioutil.ReadAll(file)

		if err != nil {
			util.WriteBadRequestResponse(w, err)
			return
		}

		r.WithContext(context.WithValue(context.Background(), fe.Field, &buffer))

		next.ServeHTTP(w, r)
	})
}
