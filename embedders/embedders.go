package embedders

import (
	"errors"
	"image"
)

type Vector []float64

var (
	ErrEmptyImage = errors.New("the image is nil or empty")
)

type ImageEmbedder interface {
	Img2Vec(image.Image) (Vector, error)
	// Dims returns the number of dimensions of vectors produced by the ImageEmbedder
	Dims() int
}
