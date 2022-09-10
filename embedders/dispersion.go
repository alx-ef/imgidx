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
	rgbaImage := ImageToRGBA(image)
	var r, g, b float64
	for i := 0; i < len(rgbaImage.Pix); i += 4 {

		r += math.Abs(means[0] - float64(rgbaImage.Pix[i])/255)
		g += math.Abs(means[1] - float64(rgbaImage.Pix[i+1])/255)
		b += math.Abs(means[2] - float64(rgbaImage.Pix[i+2])/255)
		// rgbaImage.Pix[i+3] is alpha channel, it's intentionally ignored
	}
	pixelCount := float64(image.Bounds().Dx() * image.Bounds().Dy())
	return []float64{2 * r / pixelCount, 2 * g / pixelCount, 2 * b / pixelCount}, nil
}
