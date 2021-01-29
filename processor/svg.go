package processor

import (
	"context"
	"fmt"
	"github.com/aperture147/mediaproxy/util"
	"log"
)

type SvgResult struct {
	// Object that contains the SVG minify error
	ConvertError error

	// Buffer pointer that point to the SVG
	Buffer *[]byte

	// This is used to signal other goroutine
	// which is waiting for the processed result
	context.Context
	// CancelFunc Shouldn't be called outside of the processor
	Cancel context.CancelFunc
}

type Svg struct {
	Data   *[]byte
	Result *SvgResult
}

type SvgProcessorOptions struct {
	// Number of routines/threads should be run
	Routines int
}

type SvgProcessor struct {
	Queue chan *Svg

	// Make use of context pattern
	// This is used for inter-process cancelling method
	context.Context
	Cancel context.CancelFunc

	SvgProcessorOptions
}

func NewSvgProcessor(parentCtx context.Context, options SvgProcessorOptions) SvgProcessor {
	ctx, cancel := func() (context.Context, context.CancelFunc) {
		ctx := context.Background()
		if parentCtx != nil {
			ctx = parentCtx
		}
		return context.WithCancel(ctx)
	}()
	return SvgProcessor{
		Context: ctx,
		Cancel:  cancel,
		Queue:   make(chan *Svg, 10), // is 10 too much?
		SvgProcessorOptions: SvgProcessorOptions{
			Routines: func() int {
				if options.Routines != 0 {
					return options.Routines
				}
				return DefaultRoutines
			}(),
		},
	}
}

func (p *SvgProcessor) AddSvg(data *[]byte) (*SvgResult, error) {
	ctx, err := context.WithCancel(context.Background())
	result := &SvgResult{
		Buffer:  data,
		Context: ctx,
		Cancel:  err,
	}
	p.Queue <- &Svg{
		Data:   data,
		Result: result,
	}
	return result, nil
}

func (p *SvgProcessor) Run() {
	for {
		select {
		case <-p.Done():
			log.Println("Svg Processor stopped")
			return
		case svg := <-p.Queue:
			result, err := util.MinifySvg(svg.Data)
			if err != nil {
				svg.Result.ConvertError = fmt.Errorf("svg: %v", err)
				continue
			}
			svg.Result.Buffer = result
			svg.Result.Cancel()
		}
	}
}
