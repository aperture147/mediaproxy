package processor

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"
	"time"
)

func TestImageProcessor(t *testing.T) {
	p := NewImageProcessor(context.Background(), ImageProcessorOptions{
		MaxImageSize: 3840 * 1.5,
	})

	p.Start()

	inputBuf, err := ioutil.ReadFile("2.jpeg")
	if err != nil {
		t.Fatal("cannot read image 1")
	}

	image, err := p.AddImage(&inputBuf, nil)
	if err != nil {
		t.Fatal("cannot add image 1")
	}

	select {
	case <-time.After(5 * time.Second):
		t.Fatal("overslept")
	case <-image.Done():
		if image.TransformationError != nil {
			t.Fatal(image.TransformationError)
		}
		err = ioutil.WriteFile(fmt.Sprintf("%d.jpeg", time.Now().UnixNano()), *image.Buffer, 0755)
		if err != nil {
			t.Fatal(err)
		}
	}

	p.Cancel()
}

func TestImageProcessorInterrupted(t *testing.T) {

	p := NewImageProcessor(context.Background(), ImageProcessorOptions{
		MaxImageSize: 100,
	})

	p.Start()

	inputBuf, err := ioutil.ReadFile("1.jpeg")
	if err != nil {
		t.Fatal("cannot read image 1")
	}

	image, err := p.AddImage(&inputBuf, nil)
	if err != nil {
		t.Fatal("cannot add image 1")
	}

	p.Cancel()

	select {
	case <-time.After(5 * time.Second):
		t.Fatal("overslept")
	case <-image.Done():
		if image.TransformationError != nil {
			t.Fatal(image.TransformationError)
		}
		err = ioutil.WriteFile(fmt.Sprintf("%d.jpeg", time.Now().UnixNano()), *image.Buffer, 0755)
		if err != nil {
			t.Fatal(err)
		}
	}
}
