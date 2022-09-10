package embedders

import "image"

type aspectRatioEmbedder struct{}

// NewAspectRatioEmbedder returns an embedder that takes an image's aspect ratio and
// embeds it into a vector of one dimension in the range [-1, 1]
// Horizontal images have are converted into a positive value, vertical images are converted into a negative value,
// square images have value of 0.
// The bigger the difference between the images high and width, the higher the value by module
func NewAspectRatioEmbedder() ImageEmbedder {
	return aspectRatioEmbedder{}
}

func (r aspectRatioEmbedder) Img2Vec(image *image.RGBA) (Vector, error) {
	if image == nil {
		return nil, ErrEmptyImage
	}
	size := image.Bounds().Size()

	switch {
	case size.X == 0 || size.Y == 0:
		return nil, ErrEmptyImage
	case size.X == size.Y:
		return Vector{0}, nil
	case size.X > size.Y:
		return Vector{1 - float64(size.Y)/float64(size.X)}, nil
	case size.X < size.Y:
		return Vector{float64(size.X)/float64(size.Y) - 1}, nil
	}
	panic("unreachable")
}

func (r aspectRatioEmbedder) Dims() int {
	return 1
}
