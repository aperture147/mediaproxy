package processor

import (
	"context"
	"errors"
	"github.com/discord/lilliput"
	"log"
	"math"
)

const (
	// Default 3 routines to handle the job
	DefaultRoutines = 3
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
	DefaultBufferSize = 50
	// Default size (4K)
	DefaultMaxImageSize = 3840
)

type ImageType string

const (
	ImageTypeJpeg ImageType = ".jpeg"
	ImageTypePng  ImageType = ".png"
	ImageTypeWebp ImageType = ".webp"
)

type ImageQuality string

const (
	ImageQualityAvatar       ImageQuality = "avatar"       // use for avatar upload, small scaled
	ImageQualityOrganization ImageQuality = "organization" // use for upload organization image, crop down to the largest size available
	ImageQualityDefault      ImageQuality = "default"      // default upload option
	ImageQualityCustom       ImageQuality = "custom"       // custom defined image options
)

func NormalizeSizeByScaleFactor(origWidth, origHeight, maxSize int, scaleFactor float64) (int, int, error) {
	divisor := float64(maxSize) * scaleFactor
	floatWidth, floatHeight := float64(origWidth), float64(origHeight)
	ratio := math.Max(floatWidth/divisor, floatHeight/divisor)
	newWidth, newHeight := floatWidth/ratio, floatHeight/ratio
	if IsSizeValidFloat64(newWidth, newHeight, float64(maxSize), scaleFactor) {
		return int(newWidth), int(newHeight), nil
	}
	return 0, 0, errors.New("image too large")
}

// This function is called on function that doesnt casting width height to other number type
func IsSizeValidInt(width, height, maxSize int, scaleFactor float64) bool {
	return (float64(width)/float64(maxSize)/scaleFactor <= 0.8) || (float64(height)/float64(maxSize)/scaleFactor <= 0.8)
}

func IsSizeValidFloat64(width, height, maxSize, scaleFactor float64) bool {
	return (width/maxSize/scaleFactor <= 0.2) || (height/maxSize/scaleFactor <= 0.8)
}

// A map contains multiple functions that would help resizing image
// This is one of the best thing that a compiled language like Go can do
var ImageOptionsGenerator = map[ImageQuality]func(_ *lilliput.ImageHeader, maxSize int) (*ImageOptions, error){
	// avatar
	ImageQualityAvatar: func(_ *lilliput.ImageHeader, maxSize int) (*ImageOptions, error) {
		return &ImageOptions{
			ImageType: ImageTypeJpeg,
			Width:     150,
			Height:    150,
			Resize:    true,
		}, nil
	},
	// organization:
	// Try to preserve the size of image
	ImageQualityOrganization: func(header *lilliput.ImageHeader, maxSize int) (*ImageOptions, error) {
		if (header.Width() <= maxSize) && (header.Height() <= maxSize) {
			if IsSizeValidInt(header.Width(), header.Height(), maxSize, 1) {
				return &ImageOptions{
					ImageType: ImageTypeJpeg,
					Width:     header.Width(),
					Height:    header.Height(),
					Resize:    true,
				}, nil
			}
			return nil, errors.New("image too large")
		}

		width, height, err := NormalizeSizeByScaleFactor(header.Width(), header.Height(), maxSize, 1.0)
		if err != nil {
			return nil, err
		}
		return &ImageOptions{
			ImageType: ImageTypeJpeg,
			Width:     width,
			Height:    height,
			Resize:    true,
		}, nil
	},
	// default:
	// Cut the size in half. Btw same code as organization, different factor
	// There might be a lot of change in the future so this code maybe differ from
	// organization code, but what ever, this is doing good rn
	ImageQualityDefault: func(header *lilliput.ImageHeader, maxSize int) (*ImageOptions, error) {
		newWidth, newHeight := header.Width()/2, header.Height()/2
		if (header.Width() <= maxSize) && (header.Height() <= maxSize) {
			if IsSizeValidInt(newWidth, newHeight, maxSize, 1) {
				return &ImageOptions{
					ImageType: ImageTypeJpeg,
					Width:     newWidth,
					Height:    newWidth,
					Resize:    true,
				}, nil
			}
			return nil, errors.New("image too large")
		}

		width, height, err := NormalizeSizeByScaleFactor(header.Width(), header.Height(), maxSize, 0.5)
		if err != nil {
			return nil, err
		}
		return &ImageOptions{
			ImageType: ImageTypeJpeg,
			Width:     width,
			Height:    height,
			Resize:    true,
		}, nil
	},
}

var EncodeOptions = map[ImageType]map[int]int{
	ImageTypeJpeg: {lilliput.JpegQuality: 80},
	ImageTypePng:  {lilliput.PngCompression: 7},
	ImageTypeWebp: {lilliput.WebpQuality: 85},
}

var ImageQueue = make(chan *Image, 500)

type ImageResult struct {
	// Object that contains the image transformation error
	TransformationError error

	// Buffer pointer that point to the image data
	Buffer *[]byte

	context.Context
	// CancelFunc Shouldn't be called outside of the processor
	Cancel context.CancelFunc
}

type ImageOptions struct {
	ImageType ImageType
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

func (p *ImageProcessor) AddImage(buf *[]byte, opts *ImageOptions) (*ImageResult, error) {
	data, err := lilliput.NewDecoder(*buf)
	if err != nil {
		return nil, err
	}
	// Check file header to ensure that the file is ok
	header, err := data.Header()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())

	image := &Image{
		Data:   &data,
		Header: header,
		Result: &ImageResult{
			Context: ctx,
			Cancel:  cancel,
		},
		ImageOptions: opts,
	}
	ImageQueue <- image

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
	var err error
	for {
		select {
		case <-p.Done():
			p.Ops.Close()
			log.Println("Image Processor stopped")
			return
		case image := <-ImageQueue:
			buffer := make([]byte, DefaultBufferSize*1024*1024)
			imgOpts := image.ImageOptions
			// Small check to ensure that people will not put null options
			if imgOpts == nil {
				imgOpts, err = ImageOptionsGenerator[ImageQualityDefault](image.Header, p.MaxImageSize)
				if err != nil {
					image.Result.TransformationError = err
					image.Result.Cancel()
					break
				}
			}
			opts := &lilliput.ImageOptions{
				FileType: string(imgOpts.ImageType),
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
			buffer, err = p.Ops.Transform(*image.Data, opts, buffer)
			if err != nil {
				image.Result.TransformationError = err
			} else {
				image.Result.Buffer = &buffer
			}
			p.Ops.Clear()
			image.Result.Cancel()
		}
	}
}
