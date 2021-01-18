package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/discord/lilliput"
	"github.com/gorilla/mux"
	"io/ioutil"
	"mediaproxy/processor"
	"mediaproxy/server/middleware"
	"mediaproxy/util"
	"net/http"
	"time"
)

const ImageFileField = "imageFile"

const (
	ImageOptionsKey = "options"
	ImageDataKey    = "data"
)

func NewImageRouter(ctx context.Context, maxSize int) *mux.Router {
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
			util.WriteServerErrorResponse(w, errors.New("timed out"))
		case <-result.Done():
			if result.TransformationError != nil {
				util.WriteServerErrorResponse(w, err)
				return
			}
			imgBuf := result.Buffer
			hashString := util.GetMd5String(imgBuf)

			err = ioutil.WriteFile(fmt.Sprintf("%s.jpeg", hashString), *imgBuf, 0755)
			if err != nil {
				util.WriteServerErrorResponse(w, err)
				return
			}
		}
	})

	return r
}
