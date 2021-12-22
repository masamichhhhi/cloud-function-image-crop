package p

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"strconv"

	"github.com/disintegration/imaging"
)

func ResizeImage(w http.ResponseWriter, r *http.Request) {
	// parse url
	p, err := ParseQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	img, err := FetchAndResizeImage(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	encoded, err := EncodeImageToJpg(img)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(encoded.Len()))

	_, err = io.Copy(w, encoded)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type ResizerParams struct {
	url    string
	height int
	width  int
}

func NewResizerParams(url string, height, width int) ResizerParams {
	return ResizerParams{url, height, width}
}

func ParseQuery(r *http.Request) (*ResizerParams, error) {
	var p ResizerParams
	query := r.URL.Query()
	url := query.Get("url")
	if url == "" {
		return &p, errors.New("Url Param 'url' is missing")
	}

	width, _ := strconv.Atoi(query.Get("width"))
	height, _ := strconv.Atoi(query.Get("height"))

	if width == 0 && height == 0 {
		return &p, errors.New("Url Param 'height' or 'width' must be set")
	}

	p = NewResizerParams(url, height, width)

	return &p, nil
}

func FetchAndResizeImage(p *ResizerParams) (*image.Image, error) {
	var dst image.Image

	response, err := http.Get(p.url)
	if err != nil {
		return &dst, err
	}
	defer response.Body.Close()

	src, _, err := image.Decode(response.Body)
	if err != nil {
		return &dst, err
	}

	dst = imaging.Resize(src, p.width, p.height, imaging.Lanczos)

	return &dst, nil
}

func EncodeImageToJpg(img *image.Image) (*bytes.Buffer, error) {
	encoded := &bytes.Buffer{}
	err := jpeg.Encode(encoded, *img, nil)

	return encoded, err
}
