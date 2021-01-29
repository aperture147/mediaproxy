package util

import (
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/svg"
)

const SvgMimeType = "image/svg+xml"

var minifier = minify.New()

func init() {
	minifier.AddFunc(SvgMimeType, svg.Minify)
}

func MinifySvg(data *[]byte) (*[]byte, error) {
	result, err := minifier.Bytes(SvgMimeType, *data)
	return &result, err
}
