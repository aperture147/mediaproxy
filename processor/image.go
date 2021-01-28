package processor

import (
	"context"
	"errors"
	"fmt"
	"github.com/discord/lilliput"
	"log"
)

// Default 3 routines to handle the job
const DefaultRoutines = 3

const (
	// Default buffer size in MiB
	// Please note that 50 MiB is not an optimal value one, there are
	// some edge case that this buffer size is not large enough.
	// Calculating buffer size to hold the data in image compression
	// is very complex. There might be a maximum possible size for a
	// given resolution, but it's going to be extremely large.
	// If you encountered error: "buffer too small to hold image" then just
	// increase the buffer size or limit the allowed upload image size.
	//
	// Related issue: https://github.com/discord/lilliput/issues/38
	DefaultImageBufferSize = 50
	// Default size (4K)
	DefaultMaxImageSize = 3840
)

const (
	ImageTypeJpeg = "jpeg"
	ImageTypePng  = "png"
	ImageTypeWebp = "webp"
)

var EncodeOptions = map[string]map[int]int{
	ImageTypeJpeg: {lilliput.JpegQuality: 75},
	ImageTypePng:  {lilliput.PngCompression: 7},
	ImageTypeWebp: {lilliput.WebpQuality: 85},
}

var (
	ErrTransformationError = errors.New("cannot transform the image")
	ErrNilImageOptions     = errors.New("nil image options")
)

type ImageResult struct {
	// Object that contains the image transformation error
	TransformationError error

	// Buffer pointer that point to the image data
	Buffer *[]byte

	// This is used to signal other goroutine
	// which is waiting for the processed result
	context.Context
	// CancelFunc Shouldn't be called outside of the processor
	Cancel context.CancelFunc
}

type ImageOptions struct {
	ImageType string
	Width     int
	Height    int
	Resize    bool
}

type Image struct {
	Data   *lilliput.Decoder
	Header *lilliput.ImageHeader

	// Image options, use pointer for reusing purpose
	ImageOptions *ImageOptions

	// Contains the result of the image
	Result *ImageResult
}

func (p *ImageProcessor) AddImage(data *lilliput.Decoder, opts *ImageOptions) (*ImageResult, error) {
	header, _ := (*data).Header()
	ctx, cancel := context.WithCancel(context.Background())

	image := &Image{
		Data:   data,
		Header: header,
		Result: &ImageResult{
			Context: ctx,
			Cancel:  cancel,
		},
		ImageOptions: opts,
	}
	p.Queue <- image

	return image.Result, nil
}

type ImageProcessor struct {
	// ImageOps, do the image transform job
	Ops *lilliput.ImageOps

	// Simple buffered queue for multiple the processor
	Queue chan *Image

	// Make use of context pattern
	// This is used for inter-process cancelling method
	context.Context
	Cancel context.CancelFunc

	ImageProcessorOptions
}

type ImageProcessorOptions struct {
	// Max width and height size
	MaxImageSize int

	// Number of routines/threads should be run
	Routines int
}

func NewImageProcessor(parentCtx context.Context, options ImageProcessorOptions) ImageProcessor {
	ctx, cancel := func() (context.Context, context.CancelFunc) {
		ctx := context.Background()
		if parentCtx != nil {
			ctx = parentCtx
		}
		return context.WithCancel(ctx)
	}()
	return ImageProcessor{
		Ops:     lilliput.NewImageOps(options.MaxImageSize),
		Context: ctx,
		Cancel:  cancel,
		Queue:   make(chan *Image, 500), // 500 is big enough?
		ImageProcessorOptions: ImageProcessorOptions{
			MaxImageSize: func() int {
				if options.MaxImageSize != 0 {
					return options.MaxImageSize
				}
				return DefaultMaxImageSize
			}(),
			Routines: func() int {
				if options.Routines != 0 {
					return options.Routines
				}
				return DefaultRoutines
			}(),
		},
	}
}

func (p *ImageProcessor) Start() {
	log.Printf("Starting %d routines\n", p.Routines)
	for i := 1; i <= p.Routines; i++ {
		go p.Run()
	}
}

func (p *ImageProcessor) Run() {
	for {
		select {
		case <-p.Done():
			p.Ops.Close()
			log.Println("Image Processor stopped")
			return
		case image := <-p.Queue:
			buffer := make([]byte, DefaultImageBufferSize*1024*1024)
			imgOpts := image.ImageOptions
			// Small check to ensure that people will not put null options
			if imgOpts != nil {
				opts := &lilliput.ImageOptions{
					FileType: imgOpts.ImageType,
					Width:    imgOpts.Width,
					Height:   imgOpts.Height,
					ResizeMethod: func() lilliput.ImageOpsSizeMethod {
						if imgOpts.Resize {
							return lilliput.ImageOpsFit
						}
						return lilliput.ImageOpsNoResize
					}(),
					EncodeOptions: EncodeOptions[imgOpts.ImageType],
				}
				resultBuffer, err := p.Ops.Transform(*image.Data, opts, buffer)
				if err != nil {
					image.Result.TransformationError = fmt.Errorf("transformation: %v", fmt.Errorf("%v: %v", ErrTransformationError, err))
				} else {
					image.Result.Buffer = &resultBuffer
				}
				p.Ops.Clear()
			} else {
				image.Result.TransformationError = fmt.Errorf("options: %v", ErrNilImageOptions)
			}
			image.Result.Cancel()
		}
	}
}
