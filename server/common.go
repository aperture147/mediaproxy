package server

import (
	"errors"
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
