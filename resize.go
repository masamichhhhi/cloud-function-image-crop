package p

import (
	"bytes"
	"context"
	"errors"
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

	img, err := Resize(reader, width, height)
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

func ParseWidthAndHeight(r *http.Request) (int, int, error) {
	query := r.URL.Query()
	width, _ := strconv.Atoi(query.Get("w"))
	height, _ := strconv.Atoi(query.Get("h"))

	if width == 0 && height == 0 {
		return 0, 0, errors.New("Url Param 'h' or 'w' must be set")
	}

	return width, height, nil
}

func Resize(r io.Reader, w, h int) (*image.Image, error) {
	var dst image.Image

	src, _, err := image.Decode(r)
	if err != nil {
		return &dst, err
	}

	dst = imaging.Resize(src, w, h, imaging.Lanczos)

	return &dst, nil

}

func EncodeImageToJpg(img *image.Image) (*bytes.Buffer, error) {
	encoded := &bytes.Buffer{}
	err := jpeg.Encode(encoded, *img, nil)

	return encoded, err
}
