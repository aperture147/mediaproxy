package server

import (
	"context"
	"fmt"
	"github.com/aperture147/mediaproxy/processor"
	"github.com/aperture147/mediaproxy/server/middleware"
	"github.com/aperture147/mediaproxy/storage"
	"github.com/aperture147/mediaproxy/util"
	"github.com/discord/lilliput"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

const (
	ImageFileField  = "imageFile"
	ImageOptionsKey = "options"
	ImageDataKey    = "data"
)

/*
ctx: golang context
maxFileSize: max allowed file size
maxImageSize: max width and height of the image
storage: storage component, with API helps storing image
*/
func NewImageRouter(ctx context.Context, maxFileSize, maxImageSize int, storage storage.Storage) *mux.Router {
	auth := middleware.NewTokenAuthenticator()
	extractor := middleware.NewFileExtractor(maxFileSize, ImageFileField)
	decoder := middleware.NewImageDecoder(maxImageSize, ImageFileField, ImageOptionsKey, ImageDataKey)

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
			fullPath, err2 := storage.Save(hashString, "image/"+optsPtr.ImageType, imgBuf)
			if err2 != nil {
				util.WriteServerErrorResponse(w, err2)
				return
			}
			util.WriteOkResponse(w, fullPath)
		}
	})

	return r
}
