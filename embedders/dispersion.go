package embedders

import (
	"image"
	"math"
)

type colorDispersionEmbedder struct{}

// NewColorDispersionEmbedder calculates dispersion of each color in the image
// and returns a vector of 3 components: red, green, blue color dispersion in range [0..1]
func NewColorDispersionEmbedder() ImageEmbedder {
	return colorDispersionEmbedder{}
}

func (v colorDispersionEmbedder) Dims() int {
	return 3 // red, green, blue
}

func (v colorDispersionEmbedder) Img2Vec(image image.Image) (Vector, error) {
	means, err := lowResolutionEmbedder{1, 1}.Img2Vec(image)
	if err != nil {
		return nil, err
	}

	var sum_r, sum_g, sum_b float64
	for x := 0; x < image.Bounds().Dx(); x++ {
		for y := 0; y < image.Bounds().Dy(); y++ {
			r, g, b, _ := image.At(x, y).RGBA()
			sum_r += math.Abs(means[0] - float64(r>>8)/255)
			sum_g += math.Abs(means[1] - float64(g>>8)/255)
			sum_b += math.Abs(means[2] - float64(b>>8)/255)
		}
	}
	pixelCount := float64(image.Bounds().Dx() * image.Bounds().Dy())
	return []float64{2 * sum_r / pixelCount, 2 * sum_g / pixelCount, 2 * sum_b / pixelCount}, nil
}
