package router

import (
	"context"
	"errors"
	"github.com/aperture147/mediaproxy/storage"
	"os"
)

var ErrTimedOut = errors.New("timed out")

type PathResponse struct {
	Path string `json:"path"`
	Url  string `json:"url"`
}

func GetResponse(path string) PathResponse {
	return PathResponse{
		Path: path,
		Url:  os.Getenv("HOST") + path,
	}
}

type Setting struct {
	Context context.Context // father context
	Storage storage.Storage // storage component
	Path    string          // router path
}
