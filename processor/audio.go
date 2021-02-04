package processor

import (
	"context"
	"github.com/aperture147/mediaproxy/util"
	"log"
)

const (
	// Default buffer size in MiB
	// Typically an audio uploaded to the system has the length approx. 5 mins
	// It could be an ambient music or presentation talk.
	// We can easily calculated the buffer size of a 5 mins 128k bit rate mp3 file
	// would have size of approx 4.8MB or 4.58 MiB. Rounded to 5MiB
	DefaultAudioBufferSize = 5
)

type AudioResult struct {
	// Object that contains the audio conversion error
	ConvertError error

	// Buffer pointer that point to the audio
	Buffer *[]byte

	// This is used to signal other goroutine
	// which is waiting for the processed result
	context.Context
	// CancelFunc Shouldn't be called outside of the processor
	Cancel context.CancelFunc
}

type Audio struct {
	Data   *[]byte
	Result *AudioResult

	// currently this processor is hard coded to down-sample any audio
	// to 128k mp3
	// TODO: Allow user to pass some Audio processing options
}

type AudioProcessorOptions struct {
	// Number of routines/threads should be run
	Routines int
}

type AudioProcessor struct {
	Queue chan *Audio

	// Make use of context pattern
	// This is used for inter-process cancelling method
	context.Context
	Cancel context.CancelFunc

	AudioProcessorOptions
}

func (p *AudioProcessor) Start() {
	log.Printf("Starting %d routines\n", p.Routines)
	for i := 1; i <= p.Routines; i++ {
		go p.Run()
	}
}

func NewAudioProcessor(parentCtx context.Context, options AudioProcessorOptions) AudioProcessor {
	ctx, cancel := func() (context.Context, context.CancelFunc) {
		ctx := context.Background()
		if parentCtx != nil {
			ctx = parentCtx
		}
		return context.WithCancel(ctx)
	}()
	p := AudioProcessor{
		Context: ctx,
		Cancel:  cancel,
		Queue:   make(chan *Audio, 10), // is 10 too much?
		AudioProcessorOptions: AudioProcessorOptions{
			Routines: func() int {
				if options.Routines != 0 {
					return options.Routines
				}
				return DefaultRoutines
			}(),
		},
	}
	p.Start()
	return p
}

func (p *AudioProcessor) AddAudio(data *[]byte) (*AudioResult, error) {
	ctx, err := context.WithCancel(context.Background())
	result := &AudioResult{
		Buffer:  data,
		Context: ctx,
		Cancel:  err,
	}
	p.Queue <- &Audio{
		Data:   data,
		Result: result,
	}
	return result, nil
}

func (p *AudioProcessor) Run() {
	for {
		select {
		case <-p.Done():
			log.Println("Audio Processor stopped")
			return
		case audio := <-p.Queue:
			result, err := util.AudioDownSampleToMp3(audio.Data, DefaultAudioBufferSize)
			if err != nil {
				audio.Result.ConvertError = err
				continue
			}
			audio.Result.Buffer = result
			audio.Result.Cancel()
		}
	}
}
