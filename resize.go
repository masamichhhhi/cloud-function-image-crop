package p

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/disintegration/imaging"
	"google.golang.org/api/option"
)

func ResizeStorageImage(w http.ResponseWriter, r *http.Request) {
	width, height, err := ParseWidthAndHeight(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	credentialFilePath := "./key.json"
	bucketName := "image-croppper-source"

	objectPath := strings.Replace(r.URL.Path, "/", "", 1)

	ctx := context.Background()

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialFilePath))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	obj := client.Bucket(bucketName).Object(objectPath)

	reader, err := obj.NewReader(ctx)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer reader.Close()

	resizePrams := NewResizerParams(height, width)

	img, err := Resize(reader, resizePrams)
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
	height int
	width  int
}

func NewResizerParams(height, width int) ResizerParams {
	return ResizerParams{height, width}
}

func ParseWidthAndHeight(r *http.Request) (int, int, error) {
	query := r.URL.Query()
	width, _ := strconv.Atoi(query.Get("w"))
	height, _ := strconv.Atoi(query.Get("h"))

	return width, height, nil
}

func Resize(r io.Reader, params ResizerParams) (*image.Image, error) {
	var dst image.Image

	src, _, err := image.Decode(r)
	if err != nil {
		return &dst, err
	}

	if params.width == 0 && params.height == 0 {
		dst = src
	} else {
		dst = imaging.Resize(src, params.width, params.height, imaging.Lanczos)
	}

	return &dst, nil

}

func EncodeImageToJpg(img *image.Image) (*bytes.Buffer, error) {
	encoded := &bytes.Buffer{}
	err := jpeg.Encode(encoded, *img, nil)

	return encoded, err
}
