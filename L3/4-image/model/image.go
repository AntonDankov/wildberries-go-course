package model

import (
	"fmt"

	"github.com/disintegration/imaging"
)

type ImageType int

const (
	Original ImageType = iota
	Modified
)

type ImageOperations uint8

const (
	Resize ImageOperations = 1 << iota
	Watermark
)

// ImageInfo Stores in kafka
type ImageInfo struct {
	ID           string
	Extension    imaging.Format
	Type         ImageType
	Operations   ImageOperations
	ResizeWidth  int
	ResizeHeight int
}

func (image ImageInfo) GetName() string {
	return fmt.Sprintf("%s.%s", image.ID, image.Extension.String())
}

type ImageProcessingType int

const (
	Waiting ImageProcessingType = iota
	Processed
	Deleted
	Failed
)

// ImageStatus Stores in database
type ImageStatus struct {
	ID          string
	Extension   string
	ProcessType ImageProcessingType
}
