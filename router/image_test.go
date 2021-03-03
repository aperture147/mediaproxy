package router

import (
	"context"
	"github.com/aperture147/mediaproxy/storage"
	"log"
	"net/http"
	"testing"
)

func TestNewImageRouter(t *testing.T) {
	setting := ImageRouterSetting{
		Setting: Setting{
			Context: context.Background(),
			Storage: storage.NewS3Storage("h3d-content-upload", "/uploads/image"),
			Path:    "/image/upload",
		},
		MaxFileSize:     10,
		MaxImageDimSize: 3840 * 2,
	}
	log.Fatalln(http.ListenAndServe(":8080", NewImageRouter(setting)))
}
