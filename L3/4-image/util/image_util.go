package util

import (
	"bytes"
	"fmt"
	"image"
	"wildberries-go-course/L3-4/model"

	"github.com/disintegration/imaging"
)

var watermarkImage image.Image

const WatermarkOpacity = 60

func InitWatermark(path string) error {
	img, err := imaging.Open(path)
	if err != nil {
		return fmt.Errorf("failed to load watermark: %w", err)
	}
	watermarkImage = img
	return nil
}

func OperateOnImage(imageInfo model.ImageInfo, imageData []byte) ([]byte, error) {
	operatedImageData := imageData
	var err error
	if imageInfo.Operations&model.Watermark != 0 {
		operatedImageData, err = AddWatermark(operatedImageData, imageInfo.Extension)
		if err != nil {
			return nil, err
		}
	}
	if imageInfo.Operations&model.Resize != 0 {
		operatedImageData, err = ResizeImage(operatedImageData, imageInfo.Extension, imageInfo.ResizeWidth, imageInfo.ResizeHeight)
		if err != nil {
			return nil, err
		}
	}
	return operatedImageData, nil
}

func ResizeImage(imageData []byte, extension imaging.Format, width int, height int) ([]byte, error) {
	image, err := imaging.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, err
	}

	resized := imaging.Resize(image, width, height, imaging.Lanczos)

	var buf bytes.Buffer
	err = imaging.Encode(&buf, resized, extension)
	return buf.Bytes(), err
}

func AddWatermark(imageData []byte, extension imaging.Format) ([]byte, error) {
	data, err := imaging.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, err
	}

	bounds := data.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	watermarkWidth := width / 5
	watermarkMarginWidth := width / 20
	watermarkMarginHeight := height / 20
	watermark := imaging.Resize(watermarkImage, watermarkWidth, 0, imaging.Lanczos)

	watermarkBounds := watermark.Bounds()
	x := width - watermarkBounds.Dx() - watermarkMarginWidth
	y := height - watermarkBounds.Dy() - watermarkMarginHeight

	dataWithWatermark := imaging.Overlay(data, watermark, image.Pt(x, y), WatermarkOpacity)

	var buf bytes.Buffer
	err = imaging.Encode(&buf, dataWithWatermark, extension)
	return buf.Bytes(), err
}

func FormatToContentType(format imaging.Format) string {
	switch format {
	case imaging.JPEG:
		return "image/jpeg"
	case imaging.PNG:
		return "image/png"
	case imaging.GIF:
		return "image/gif"
	case imaging.BMP:
		return "image/bmp"
	case imaging.TIFF:
		return "image/tiff"
	default:
		return "application/octet-stream"
	}
}
