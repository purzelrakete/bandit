package bandit

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Opener can be used to reopen underlying file descriptors.
type Opener interface {
	Open() (io.ReadCloser, error)
}

// NewOpener returns an http opener or a file opener depending on `ref`.
func NewOpener(ref string) Opener {
	var opener Opener
	if strings.Index(ref, "http://") >= 0 {
		opener = NewHTTPOpener(ref)
	} else {
		opener = NewFileOpener(ref)
	}

	return opener
}

// NewHTTPOpener returns an opener using an underlying URL.
func NewHTTPOpener(url string) Opener {
	return &httpOpener{
		URL: url,
	}
}

type httpOpener struct {
	URL string
}

func (o *httpOpener) Open() (io.ReadCloser, error) {
	resp, err := http.Get(o.URL)
	if err != nil {
		return nil, fmt.Errorf("http GET failed: %s", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("http GET not 200: %s", resp.StatusCode)
	}
	return resp.Body, nil
}

// NewFileOpener returns an Opener using and underlying file.
func NewFileOpener(filename string) Opener {
	return &fileOpener{
		Filename: filename,
	}
}

type fileOpener struct {
	Filename string
}

func (o *fileOpener) Open() (io.ReadCloser, error) {
	reader, err := os.Open(o.Filename)
	if err != nil {
		return nil, err
	}

	return reader, err
}
