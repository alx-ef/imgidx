package embedders

import (
	"errors"
	"image"
)

type Vector []float64

var (
	ErrEmptyImage = errors.New("the image is nil or empty")
)

func (v1 Vector) Distance(v2 Vector) float64 {
	// same as for the kdtree.Point type
	var sum float64
	for dim := range v1 {
		d := v1[dim] - v2[dim]
		sum += d * d
	}
	return sum
}

type ImageEmbedder interface {
	Img2Vec(*image.RGBA) (Vector, error)
	// Dims returns the number of dimensions of vectors produced by the ImageEmbedder
	Dims() int
}
