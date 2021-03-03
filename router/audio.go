package router

import (
	"github.com/aperture147/mediaproxy/processor"
	"github.com/aperture147/mediaproxy/router/middleware"
	"github.com/aperture147/mediaproxy/util"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

const AudioFileField = "audioFile"

type AudioRouterSetting struct {
	Setting         // base setting
	MaxFileSize int // max audio file size allowed
}

func NewAudioRouter(setting AudioRouterSetting) *mux.Router {
	auth := middleware.NewTokenAuthenticator()
	extractor := middleware.NewFileExtractor(setting.MaxFileSize, AudioFileField)
	r := mux.NewRouter()
	r.Use(auth.Verify, extractor.Verify)

	opts := processor.AudioProcessorOptions{}
	p := processor.NewAudioProcessor(setting.Context, opts)

	r.HandleFunc(setting.Path, func(w http.ResponseWriter, r *http.Request) {
		dataPtr := r.Context().Value(AudioFileField).(*[]byte)

		result, err := p.AddAudio(dataPtr)
		if err != nil {
			ServerErrorResponseAndLog(w, "audio add failed", ErrAddToProcessor)
			return
		}
		select {
		case <-time.After(30 * time.Second):
			ServerErrorResponseAndLog(w, "convert timed out", ErrTimedOut)
		case <-result.Done():
			if result.ConvertError != nil {
				ServerErrorResponseAndLog(w, "convert failed", result.ConvertError)
				return
			}
			audioBuf := result.Buffer
			hashString := util.GetMd5String(audioBuf)

			path, err2 := setting.Storage.Save(hashString, "audio/mpeg", audioBuf)
			if err2 != nil {
				ServerErrorResponseAndLog(w, "audio save failed", result.ConvertError)
				return
			}
			util.WriteOkResponse(w, GetResponse(path))
		}
	})
	return r
}
