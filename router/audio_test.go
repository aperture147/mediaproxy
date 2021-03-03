package router

import (
	"context"
	"github.com/aperture147/mediaproxy/storage"
	"log"
	"net/http"
	"testing"
)

func TestNewAudioRouter(t *testing.T) {
	setting := AudioRouterSetting{
		Setting: Setting{
			Context: context.Background(),
			Storage: storage.NewS3Storage("h3d-content-upload", "/uploads/audio"),
			Path:    "/audio/upload",
		},
		MaxFileSize: 0,
	}
	log.Fatalln(http.ListenAndServe(":8080", NewAudioRouter(setting)))
}
