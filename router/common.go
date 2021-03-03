package router

import (
	"context"
	"errors"
	"fmt"
	"github.com/aperture147/mediaproxy/storage"
	"github.com/aperture147/mediaproxy/util"
	"log"
	"net/http"
	"os"
)

var (
	ErrTimedOut       = errors.New("timed out")
	ErrAddToProcessor = errors.New("cannot add to processor")
)

type PathResponse struct {
	Path string `json:"path"`
	Url  string `json:"url"`
}

func GetResponse(path string) PathResponse {
	return PathResponse{
		Path: path,
		Url:  os.Getenv("CDN_HOST") + path,
	}
}

func ServerErrorResponseAndLog(w http.ResponseWriter, msg string, err error) {
	newErr := fmt.Errorf("%s: %v", msg, ErrAddToProcessor)
	log.Println(newErr)
	util.WriteServerErrorResponse(w, newErr)
}

type Setting struct {
	Context context.Context // father context
	Storage storage.Storage // storage component
	Path    string          // router path
}
