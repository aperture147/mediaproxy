package router

import (
	"fmt"
	"github.com/aperture147/mediaproxy/processor"
	"github.com/aperture147/mediaproxy/router/middleware"
	"github.com/aperture147/mediaproxy/util"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

const AudioFileField = "audioFile"

type AudioRouterSetting struct {
	Setting             // base setting
	MaxFileSize     int // max audio file size allowed
	MaxImageDimSize int // max image width and height
}

func NewAudioRouter(setting AudioRouterSetting) *mux.Router {
	auth := middleware.NewTokenAuthenticator()
	extractor := middleware.NewFileExtractor(setting.MaxFileSize, AudioFileField)
	r := mux.NewRouter()
	r.Use(auth.Verify, extractor.Verify)

	opts := processor.AudioProcessorOptions{}
	p := processor.NewAudioProcessor(setting.Context, opts)

	r.HandleFunc("/audio/upload", func(w http.ResponseWriter, r *http.Request) {
		dataPtr := r.Context().Value(AudioFileField).(*[]byte)

		result, err := p.AddAudio(dataPtr)
		if err != nil {
			util.WriteServerErrorResponse(w, err)
			return
		}
		select {
		case <-time.After(30 * time.Second):
			util.WriteServerErrorResponse(w, fmt.Errorf("convert: %v", ErrTimedOut))
		case <-result.Done():
			if result.ConvertError != nil {
				util.WriteServerErrorResponse(w, err)
				return
			}
			audioBuf := result.Buffer
			hashString := util.GetMd5String(audioBuf)

			path, err2 := setting.Storage.Save(hashString, "audio/mpeg", audioBuf)
			if err2 != nil {
				util.WriteServerErrorResponse(w, err2)
				return
			}
			util.WriteOkResponse(w, GetResponse(path))
		}
	})
	return r
}
