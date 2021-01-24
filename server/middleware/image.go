package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/discord/lilliput"
	"math"
	"mediaproxy/processor"
	"mediaproxy/util"
	"net/http"
)

const ImageQualityField = "quality"

const (
	ImageQualityAvatar       = "avatar"       // use for avatar upload, small scaled
	ImageQualityOrganization = "organization" // use for upload organization image, crop down to the largest size available
	ImageQualityCustom       = "custom"       // custom defined image options
)

var (
	ErrImageSizeTooLarge      = errors.New("image size too large") // Image width and height too large
	ErrImageDecodeFailed      = errors.New("image decode failed")  // Cannot decode image
	ErrImageHeaderCheckFailed = errors.New("image header check failed")
)

func NormalizeSizeByScaleFactor(origWidth, origHeight, maxSize int, scaleFactor float64) (int, int, error) {
	divisor := float64(maxSize) * scaleFactor
	floatWidth, floatHeight := float64(origWidth), float64(origHeight)
	ratio := math.Max(floatWidth/divisor, floatHeight/divisor)
	newWidth, newHeight := floatWidth/ratio, floatHeight/ratio
	if IsSizeValidFloat64(newWidth, newHeight, float64(maxSize), scaleFactor) {
		return int(newWidth), int(newHeight), nil
	}
	return 0, 0, fmt.Errorf("size normaliziation: %v", ErrImageSizeTooLarge)
}

// This function is called on function that doesnt casting width height to other number type
func IsSizeValidInt(width, height, maxSize int, scaleFactor float64) bool {
	return (float64(width)/float64(maxSize)/scaleFactor <= 0.8) || (float64(height)/float64(maxSize)/scaleFactor <= 0.8)
}

func IsSizeValidFloat64(width, height, maxSize, scaleFactor float64) bool {
	return (width/maxSize/scaleFactor <= 0.2) || (height/maxSize/scaleFactor <= 0.8)
}

func ImageQualityAvatarGenerator(_ *lilliput.ImageHeader, _ int) (*processor.ImageOptions, error) {
	return &processor.ImageOptions{
		ImageType: processor.ImageTypeJpeg,
		Width:     150,
		Height:    150,
		Resize:    true,
	}, nil
}

func ImageQualityOrganizationGenerator(header *lilliput.ImageHeader, maxSize int) (*processor.ImageOptions, error) {
	if (header.Width() <= maxSize) && (header.Height() <= maxSize) {
		if IsSizeValidInt(header.Width(), header.Height(), maxSize, 1) {
			return &processor.ImageOptions{
				ImageType: processor.ImageTypeJpeg,
				Width:     header.Width(),
				Height:    header.Height(),
				Resize:    true,
			}, nil
		}
		return nil, ErrImageSizeTooLarge
	}

	width, height, err := NormalizeSizeByScaleFactor(header.Width(), header.Height(), maxSize, 1.0)
	if err != nil {
		return nil, err
	}
	return &processor.ImageOptions{
		ImageType: processor.ImageTypeJpeg,
		Width:     width,
		Height:    height,
		Resize:    true,
	}, nil
}

func ImageQualityDefaultGenerator(header *lilliput.ImageHeader, maxSize int) (*processor.ImageOptions, error) {
	newWidth, newHeight := header.Width()/2, header.Height()/2
	if (header.Width() <= maxSize) && (header.Height() <= maxSize) {
		if IsSizeValidInt(newWidth, newHeight, maxSize, 1) {
			return &processor.ImageOptions{
				ImageType: processor.ImageTypeJpeg,
				Width:     newWidth,
				Height:    newWidth,
				Resize:    true,
			}, nil
		}
		return nil, ErrImageSizeTooLarge
	}

	width, height, err := NormalizeSizeByScaleFactor(header.Width(), header.Height(), maxSize, 0.5)
	if err != nil {
		return nil, err
	}
	return &processor.ImageOptions{
		ImageType: processor.ImageTypeJpeg,
		Width:     width,
		Height:    height,
		Resize:    true,
	}, nil
}

// A function generates image options
func ImageOptionsGenerator(header *lilliput.ImageHeader, quality string, maxSize int) (*processor.ImageOptions, error) {
	switch quality {
	// avatar
	// cut the image from the center, take 150px then return it.
	case ImageQualityAvatar:
		return ImageQualityAvatarGenerator(header, maxSize)
	// organization:
	// Try to preserve the size of image
	case ImageQualityOrganization:
		return ImageQualityOrganizationGenerator(header, maxSize)
	// default:
	// Cut the size in half. Btw same code as organization, different factor
	// There might be a lot of change in the future so this code maybe differ from
	// organization code, but what ever, this is doing good rn
	default:
		return ImageQualityDefaultGenerator(header, maxSize)
	}
}

type ImageDecoder struct {
	FileField    string
	OptionsField string
	DataField    string
	MaxSize      int
}

func NewImageDecoder(maxSize int, fileField, optsField, dataField string) ImageDecoder {
	return ImageDecoder{
		FileField:    fileField,
		OptionsField: optsField,
		DataField:    dataField,
		MaxSize:      maxSize,
	}
}

func (i ImageDecoder) Decode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bufferPtr := r.Context().Value(i.FileField).(*[]byte)
		data, err := lilliput.NewDecoder(*bufferPtr)
		if err != nil {
			util.WriteBadRequestResponse(w, fmt.Errorf("%v: %v", ErrImageDecodeFailed, err))
			return
		}

		// Check file header to ensure that the file is ok
		header, err := data.Header()
		if err != nil {
			util.WriteBadRequestResponse(w, fmt.Errorf("%v: %v", ErrImageHeaderCheckFailed, err))
			return
		}

		quality := r.FormValue(ImageQualityField)
		opts, err := ImageOptionsGenerator(header, quality, i.MaxSize)

		if err != nil {
			util.WriteServerErrorResponse(w, fmt.Errorf("image options: %v", err))
			return
		}

		ctx := context.WithValue(context.Background(), i.OptionsField, opts)
		r.WithContext(context.WithValue(ctx, i.DataField, data))

		next.ServeHTTP(w, r)
	})
}
