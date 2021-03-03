package router

import (
	"github.com/aperture147/mediaproxy/processor"
	"github.com/aperture147/mediaproxy/router/middleware"
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

type ImageRouterSetting struct {
	Setting
	MaxFileSize     int
	MaxImageDimSize int
}

/*
ctx: golang context
maxFileSize: max allowed file size
maxImageSize: max width and height of the image
storage: storage component, with API helps storing image
*/
func NewImageRouter(setting ImageRouterSetting) *mux.Router {
	auth := middleware.NewTokenAuthenticator()
	extractor := middleware.NewFileExtractor(setting.MaxFileSize, ImageFileField)
	decoder := middleware.NewImageDecoder(setting.MaxImageDimSize, ImageFileField, ImageOptionsKey, ImageDataKey)

	r := mux.NewRouter()
	r.Use(auth.Verify, extractor.Verify, decoder.Decode)

	opts := processor.ImageProcessorOptions{
		MaxImageSize: setting.MaxImageDimSize,
	}
	p := processor.NewImageProcessor(setting.Context, opts)

	r.HandleFunc(setting.Path, func(w http.ResponseWriter, r *http.Request) {
		optsPtr := r.Context().Value(ImageOptionsKey).(*processor.ImageOptions)
		dataPtr := r.Context().Value(ImageDataKey).(*lilliput.Decoder)

		result, err := p.AddImage(dataPtr, optsPtr)
		if err != nil {
			ServerErrorResponseAndLog(w, "image add failed", ErrAddToProcessor)
			return
		}
		select {
		case <-time.After(30 * time.Second):
			ServerErrorResponseAndLog(w, "transform timed out", ErrTimedOut)
		case <-result.Done():
			if result.TransformationError != nil {
				ServerErrorResponseAndLog(w, "transformation failed", result.TransformationError)
				return
			}
			imgBuf := result.Buffer
			hashString := util.GetMd5String(imgBuf)
			path, err2 := setting.Storage.Save(hashString, "image/"+optsPtr.ImageType, imgBuf)
			if err2 != nil {
				ServerErrorResponseAndLog(w, "image save failed", err2)
				return
			}
			util.WriteOkResponse(w, GetResponse(path))
		}
	})

	return r
}
