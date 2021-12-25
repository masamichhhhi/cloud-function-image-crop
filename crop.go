package p

import (
	"context"
	"image"
	"io"
	"net/http"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/disintegration/imaging"
	"google.golang.org/api/option"
)

func CropImage(w http.ResponseWriter, r *http.Request) {
	width, height, cropStartX, cropStartY, err := ParseCropParams(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	credentialFilePath := "./key.json"
	bucketName := "image-cropper-source"

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

	cropParams := NewCropParams(width, height, cropStartX, cropStartY)

	img, err := Crop(reader, cropParams)
	encoded, err := EncodeImageToJpg(img)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

type CropParams struct {
	width      int
	height     int
	cropStartX int
	cropStartY int
}

func NewCropParams(width, height, cropStartX, cropStartY int) CropParams {
	return CropParams{width, height, cropStartX, cropStartY}
}

func ParseCropParams(r *http.Request) (int, int, int, int, error) {
	query := r.URL.Query()
	width, _ := strconv.Atoi(query.Get("w"))
	height, _ := strconv.Atoi(query.Get("h"))

	cropStartX, _ := strconv.Atoi(query.Get("cropStartX"))
	cropStartY, _ := strconv.Atoi(query.Get("cropStartY"))

	return width, height, cropStartX, cropStartY, nil
}

func Crop(r io.Reader, params CropParams) (*image.Image, error) {
	var dst image.Image

	src, _, err := image.Decode(r)
	if err != nil {
		return &dst, err
	}

	if params.width == 0 && params.height == 0 {
		dst = src
	} else {
		dst = imaging.Crop(src, image.Rectangle{
			Min: image.Point{X: params.cropStartX, Y: params.cropStartY},
			Max: image.Point{X: params.cropStartX + params.width, Y: params.cropStartY + params.height},
		})
	}

	return &dst, nil
}
