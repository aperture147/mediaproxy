package server

import (
	"context"
	"fmt"
	"github.com/aperture147/mediaproxy/processor"
	"github.com/aperture147/mediaproxy/server/middleware"
	"github.com/aperture147/mediaproxy/storage"
	"github.com/aperture147/mediaproxy/util"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

const AudioFileField = "audioFile"

func NewAudioRouter(ctx context.Context, maxSize int, storage storage.Storage) *mux.Router {
	auth := middleware.NewTokenAuthenticator()
	extractor := middleware.NewFileExtractor(maxSize, AudioFileField)
	r := mux.NewRouter()
	r.Use(auth.Verify, extractor.Verify)

	opts := processor.AudioProcessorOptions{}
	p := processor.NewAudioProcessor(ctx, opts)

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

			fullPath, err2 := storage.Save(hashString, "audio/mpeg", audioBuf)
			if err2 != nil {
				util.WriteServerErrorResponse(w, err2)
				return
			}
			util.WriteOkResponse(w, fullPath)
		}
	})
	return r
}
