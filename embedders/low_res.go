package embedders

import (
	"fmt"
	"image"
)

type lowResolutionEmbedder struct {
	Width  int
	Height int
}

func (v lowResolutionEmbedder) Dims() int {
	return v.Height * v.Width * 4
}

// Img2Vec returns the vector representation of the image.
func (v lowResolutionEmbedder) Img2Vec(img image.Image) (Vector, error) {
	if v.Width <= 0 || v.Height <= 0 {
		return nil, fmt.Errorf("lowResolutionEmbedder's Width and Height parameters must be greater than 0")
	}
	if img == nil {
		return nil, fmt.Errorf("image must be non-nil")
	}
	if img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
		return nil, fmt.Errorf("image width and height must be greater than 0")
	}
	if v.Width > img.Bounds().Dx() || v.Height > img.Bounds().Dy() {
		return nil, fmt.Errorf(
			"image width and height must not be less than lowResolutionEmbedder's Width and Height parameters")
	}
	vec := make(Vector, v.Height*v.Width*4)
	for row := 0; row < v.Height; row++ {
		for col := 0; col < v.Width; col++ {
			rgba := getAverageColor(img,
				col*img.Bounds().Dx()/v.Width,
				(col+1)*img.Bounds().Dx()/v.Width,
				row*img.Bounds().Dy()/v.Height,
				(row+1)*img.Bounds().Dy()/v.Height,
			)
			for i, f := range rgba {
				vec[row*v.Width*4+col*4+i] = f
			}
		}
	}
	return vec, nil
}

// getAverageColor returns the average color of the given area of the image.
func getAverageColor(img image.Image, minX int, maxX int, minY int, maxY int) (rgba [4]float64) {
	if rgba, ok := img.(*image.RGBA); ok {
		return getAverageColorRGBA(*rgba, minX, maxX, minY, maxY)
	}
	for i := minX; i < maxX; i++ {
		for k := minY; k < maxY; k++ {
			r, g, b, a := img.At(i, k).RGBA()
			rgba[0] += float64(r >> 8)
			rgba[1] += float64(g >> 8)
			rgba[2] += float64(b >> 8)
			rgba[3] += float64(a >> 8)
		}
	}

	pixelCount := float64((maxX - minX) * (maxY - minY))
	for i := 0; i < 4; i++ {
		rgba[i] /= pixelCount * 255
	}
	return rgba
}

// getAverageColorRGBA is optimized version of getAverageColor for RGBA images.
func getAverageColorRGBA(img image.RGBA, minX int, maxX int, minY int, maxY int) [4]float64 {
	var (
		avgColorsInt   [4]int // 4 channels: R, G, B, A
		avgColorsFloat [4]float64
		pixelCount     = (maxX - minX) * (maxY - minY)
		devider        = float64(pixelCount * 255)
	)
	if devider == 0 { //image is empty, return zeros
		return avgColorsFloat
	}
	for x := minX; x < maxX; x++ {
		for y := minY; y < maxY; y++ {
			i := img.PixOffset(x, y)
			s := img.Pix[i : i+4 : i+4]
			for i, c := range s {
				avgColorsInt[i] += int(c)
			}
		}
	}
	for i, c := range avgColorsInt {
		avgColorsFloat[i] = float64(c) / devider
	}
	return avgColorsFloat
}

// NewLowResolutionEmbedder returns an embedder that takes an image, splits it into height*width rectangle
// and calculates the average amount of each channel (3 colors + alpha) in each rectangle.
// It produces vectors of height*width*4 numbers in range [0..1]
// E.g. for an image filled with red color (#FF0000) entirely and the following parameters: height=3, width=4
// the resulting vector would be [1,0,0,0,1,0,0,0 ... 1,0,0,0], total vector length is 48
func NewLowResolutionEmbedder(width int, height int) ImageEmbedder {
	return lowResolutionEmbedder{width, height}
}
