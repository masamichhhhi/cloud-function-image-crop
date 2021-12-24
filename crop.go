package p

import (
	"net/http"
	"strconv"
)

func CropImage(w http.ResponseWriter, r *http.Request) {

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
