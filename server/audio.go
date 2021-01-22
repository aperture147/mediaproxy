package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"mediaproxy/processor"
	"mediaproxy/server/middleware"
	"mediaproxy/util"
	"net/http"
	"time"
)

const AudioFileField = "audioFile"

func NewAudioRouter(ctx context.Context, maxSize int) *mux.Router {
	auth := middleware.NewTokenAuthenticator()
	extractor := middleware.NewFileExtractor(maxSize, AudioFileField)
	r := mux.NewRouter()
	r.Use(auth.Verify, extractor.Verify)

	opts := processor.AudioProcessorOptions{}
	p := processor.NewAudioProcessor(ctx, opts)

	r.HandleFunc("/image/upload", func(w http.ResponseWriter, r *http.Request) {
		dataPtr := r.Context().Value(AudioFileField).(*[]byte)

		result, err := p.AddAudio(dataPtr)
		if err != nil {
			util.WriteServerErrorResponse(w, err)
			return
		}
		select {
		case <-time.After(30 * time.Second):
			util.WriteServerErrorResponse(w, errors.New("timed out"))
		case <-result.Done():
			if result.ConvertError != nil {
				util.WriteServerErrorResponse(w, err)
				return
			}
			audioBuf := result.Buffer
			hashString := util.GetMd5String(audioBuf)

			err = ioutil.WriteFile(fmt.Sprintf("%s.mp3", hashString), *audioBuf, 0755)
			if err != nil {
				util.WriteServerErrorResponse(w, err)
				return
			}
			util.WriteOkResponse(w, nil)
		}
	})
	return r
}
