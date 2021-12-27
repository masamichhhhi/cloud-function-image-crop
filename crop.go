package p

import (
	"bytes"
	"context"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
		return
	}

	credentialFilePath := "./key.json"
	bucketName := "image-croppper-source"

	objectPath := strings.Replace(r.URL.Path, "/", "", 1)

	ctx := context.Background()

	var client *storage.Client

	if os.Getenv("ENVIRONMENT") == "production" {
		client, err = storage.NewClient(ctx)
	} else {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(credentialFilePath))
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	obj := client.Bucket(bucketName).Object(objectPath)

	reader, err := obj.NewReader(ctx)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer reader.Close()

	// TODO: パラメータのバリデーションする
	cropParams := NewCropParams(width, height, cropStartX, cropStartY)

	var encoded *bytes.Buffer
	if filepath.Ext(objectPath) == ".gif" {
		gif, err := cropGif(reader, cropParams)
		encoded, err = encodeGif(gif)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		img, err := Crop(reader, cropParams)
		encoded, err = EncodeImageToJpg(img)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Type", "image/gif")
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

func EncodeImageToJpg(img *image.Image) (*bytes.Buffer, error) {
	encoded := &bytes.Buffer{}
	err := jpeg.Encode(encoded, *img, nil)

	return encoded, err
}

func cropGif(reader io.Reader, params CropParams) (*gif.GIF, error) {
	outGif := &gif.GIF{}
	g, err := gif.DecodeAll(reader)

	if err != nil {
		return nil, err
	}

	if params.width == 0 && params.height == 0 && params.cropStartX == 0 && params.cropStartY == 0 {
		outGif = g
	} else {
		for _, img := range g.Image {
			p := image.NewPaletted(image.Rect(0, 0, params.width, params.height), img.Palette)

			cropedImage := imaging.Crop(img, image.Rectangle{
				Min: image.Point{X: params.cropStartX, Y: params.cropStartY},
				Max: image.Point{X: params.cropStartX + params.width, Y: params.cropStartY + params.height},
			})

			draw.Draw(p, image.Rect(0, 0, params.width, params.height), cropedImage, image.ZP, draw.Src)

			outGif.Image = append(outGif.Image, p)
			outGif.Delay = append(outGif.Delay, 0)
		}
	}

	return outGif, nil
}

func encodeGif(inGif *gif.GIF) (*bytes.Buffer, error) {
	encoded := &bytes.Buffer{}

	err := gif.EncodeAll(encoded, inGif)
	return encoded, err
}
