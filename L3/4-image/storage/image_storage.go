package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"wildberries-go-course/L3-4/model"
)

type ImageStorage struct {
	BasePath string
	Depth    int
	Width    int
}

func (imageStorage *ImageStorage) StoreImage(image model.ImageInfo, imageData []byte) error {
	filePath := imageStorage.getImageFilePath(image)
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	err := os.WriteFile(filePath, imageData, 0644)
	if err != nil {
		return fmt.Errorf("writing image failed: %w", err)
	}

	return nil
}

func (imageStorage *ImageStorage) GetImageData(image *model.ImageInfo) ([]byte, error) {
	filePath := imageStorage.getImageFilePath(*image)
	return os.ReadFile(filePath)
}

func (imageStorage *ImageStorage) getImageFilePath(image model.ImageInfo) string {
	parts := []string{imageStorage.BasePath}
	var typeString string
	switch image.Type {
	case model.Original:
		typeString = "original"
	case model.Modified:
		typeString = "modified"
	}
	parts = append(parts, typeString)
	for i := 0; i < imageStorage.Depth; i++ {
		start := i * imageStorage.Width
		end := start + imageStorage.Width
		if end > len(image.ID) {
			break
		}
		parts = append(parts, image.ID[start:end])
	}
	parts = append(parts, image.GetName())

	return filepath.Join(parts...)
}
