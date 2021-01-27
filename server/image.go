package server

import (
	"context"
	"fmt"
	"github.com/discord/lilliput"
	"github.com/gorilla/mux"
	"mediaproxy/processor"
	"mediaproxy/server/middleware"
	"mediaproxy/storage"
	"mediaproxy/util"
	"net/http"
	"time"
)

const (
	ImageFileField  = "imageFile"
	ImageOptionsKey = "options"
	ImageDataKey    = "data"
)

func NewImageRouter(ctx context.Context, maxSize int, storage storage.Storage) *mux.Router {
	auth := middleware.NewTokenAuthenticator()
	extractor := middleware.NewFileExtractor(maxSize, ImageFileField)
	decoder := middleware.NewImageDecoder(maxSize, ImageFileField, ImageOptionsKey, ImageDataKey)

	r := mux.NewRouter()
	r.Use(auth.Verify, extractor.Verify, decoder.Decode)

	opts := processor.ImageProcessorOptions{
		MaxImageSize: 1080,
	}
	p := processor.NewImageProcessor(ctx, opts)

	r.HandleFunc("/image/upload", func(w http.ResponseWriter, r *http.Request) {
		optsPtr := r.Context().Value(ImageOptionsKey).(*processor.ImageOptions)
		dataPtr := r.Context().Value(ImageDataKey).(*lilliput.Decoder)

		result, err := p.AddImage(dataPtr, optsPtr)
		if err != nil {
			util.WriteServerErrorResponse(w, err)
			return
		}
		select {
		case <-time.After(30 * time.Second):
			util.WriteServerErrorResponse(w, fmt.Errorf("transformation: %v", ErrTimedOut))
		case <-result.Done():
			if result.TransformationError != nil {
				util.WriteServerErrorResponse(w, err)
				return
			}
			imgBuf := result.Buffer
			hashString := util.GetMd5String(imgBuf)
			fullPath, err := storage.Save(hashString, "image/"+optsPtr.ImageType, imgBuf)
			if err != nil {
				util.WriteServerErrorResponse(w, err)
				return
			}
			util.WriteOkResponse(w, fullPath)
		}
	})

	return r
}
