package processor

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

func TestFFmpegPipe(t *testing.T) {
	file, err := os.Open("./test.mp3")
	if err != nil {
		t.Fatal(err)
	}

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("ffmpeg", "-y", // Yes to all
		"-hide_banner", "-loglevel", "panic", // Hide all logs
		"-i", "pipe:0", // take stdin
		"-map_metadata", "-1", // strip out all (mostly) metadata
		"-c:a", "libmp3lame", // use mp3 lame codec
		"-vsync", "2", // suppress "Frame rate very high for a muxer not efficiently supporting it"
		"-b:a", "128k", // Downsample to 128k
		"-f", "mp3", // using mp3 muxer
		"pipe:1", // output to stdout
	)
	resultBuffer := bytes.NewBuffer([]byte{})

	cmd.Stderr = os.Stderr
	cmd.Stdout = resultBuffer

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	err = cmd.Start()
	if err != nil {
		t.Fatal(err)
	}

	_, err = stdin.Write(buf)
	if err != nil {
		t.Fatal(err)
	}

	err = stdin.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = cmd.Wait()
	if err != nil {
		t.Fatal(err)
	}

	outputFile, err := os.Create("stdout.mp3")
	if err != nil {
		t.Fatal(err)
	}

	_, err = outputFile.Write(resultBuffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}
}
