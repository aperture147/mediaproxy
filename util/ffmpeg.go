package util

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

const (
	DefaultFFmpegPath        = "ffmpeg"
	DefaultDownSampleTimeout = 20 * time.Second
)

var ffmpegPath = DefaultFFmpegPath
var ffmpegArgs = []string{
	"-y",                                 // Yes to all
	"-hide_banner", "-loglevel", "panic", // Hide logs
	"-i", "pipe:0", // take stdin
	"-map_metadata", "-1", // strip out all (mostly) metadata
	"-c:a", "libmp3lame", // use mp3 lame codec
	"-vsync", "2", // suppress "Frame rate very high for a muxer not efficiently supporting it"
	"-b:a", "128k", // Down-sample audio bitrate to 128k
	"-f", "mp3", // using mp3 muxer
	"pipe:1", // Output audio to
}

var (
	ErrConvertError = errors.New("cannot convert audio file")
)

func GenerateError(err error) error {
	return fmt.Errorf("convert: %v", fmt.Errorf("%v: %v", ErrConvertError, err))
}

// Check this for further explaination: https://gist.github.com/aperture147/ad0f5b965912537d03b0e851bb95bd38
func AudioDownSample(buf *[]byte, allocMemSize int) (*[]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultDownSampleTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, ffmpegPath, ffmpegArgs...)

	resultBuffer := bytes.NewBuffer(make([]byte, allocMemSize*1024*1024))
	cmd.Stdout = resultBuffer

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, GenerateError(err)
	}

	err = cmd.Start()
	if err != nil {
		return nil, GenerateError(err)
	}

	_, err = stdin.Write(*buf) // pump audio data to stdin pipe
	if err != nil {
		return nil, GenerateError(err)
	}

	err = stdin.Close()
	if err != nil {
		return nil, GenerateError(err)
	}

	<-ctx.Done()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, err
	}

	result := resultBuffer.Bytes()

	return &result, nil
}
