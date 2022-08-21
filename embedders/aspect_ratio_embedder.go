package embedders

import "image"

// AspectRatioEmbedder takes an image's aspect ratio and embeds it into a vector of one dimension in the range [-1, 1]
// Horizontal images have value > 0, vertical images have aspect ratio < 0, square images have value 0
// The bigger the difference between the images high and width, the higher the value by module
type AspectRatioEmbedder struct{}

func (r AspectRatioEmbedder) Img2Vec(image image.Image) (Vector, error) {
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

func (r AspectRatioEmbedder) Dims() int {
	return 1
}
