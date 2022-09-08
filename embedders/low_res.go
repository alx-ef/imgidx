package embedders

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
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

	var (
		rgbaImg  *image.RGBA
		ycbcrImg *image.YCbCr
		ok       bool
	)

	ycbcrImg, ok = img.(*image.YCbCr)
	if !ok {
		rgbaImg = ImageToRGBA(img)
	}

	vec := make(Vector, v.Height*v.Width*4)
	for row := 0; row < v.Height; row++ {
		for col := 0; col < v.Width; col++ {
			minX := col * img.Bounds().Dx() / v.Width
			maxX := (col + 1) * img.Bounds().Dx() / v.Width
			minY := row * img.Bounds().Dy() / v.Height
			maxY := (row + 1) * img.Bounds().Dy() / v.Height
			var rgba [4]float64
			if ycbcrImg != nil {
				rgba = getAverageColorYCbCr(ycbcrImg, minX, maxX, minY, maxY)
			} else {
				rgba = getAverageColorRGBA(rgbaImg, minX, maxX, minY, maxY)
			}

			for i, f := range rgba {
				vec[row*v.Width*4+col*4+i] = f
			}
		}
	}
	return vec, nil
}

// getAverageColorRGBA returns the average color of the given area of the image.
func getAverageColorRGBA(img *image.RGBA, minX int, maxX int, minY int, maxY int) [4]float64 {
	var (
		avgColorsInt   [4]int // 4 channels: R, G, B, A
		avgColorsFloat [4]float64
		pixelCount     = (maxX - minX) * (maxY - minY)
		divider        = float64(pixelCount * 255)
	)
	if divider == 0 { //image is empty, return zeros
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
		avgColorsFloat[i] = float64(c) / divider
	}
	return avgColorsFloat
}

// getAverageColorYCbCr returns the average color of the given area of the image.
func getAverageColorYCbCr(img *image.YCbCr, minX int, maxX int, minY int, maxY int) [4]float64 {
	var (
		avgColorsInt   [4]int // 4 channels: R, G, B, A
		avgColorsFloat [4]float64
		pixelCount     = (maxX - minX) * (maxY - minY)
		divider        = float64(pixelCount * 255)
	)
	if divider == 0 { //image is empty, return zeros
		return avgColorsFloat
	}

	if divider == 0 { //image is empty, return zeros
		return avgColorsFloat
	}
	for x := minX; x < maxX; x++ {
		for y := minY; y < maxY; y++ {
			yi := img.YOffset(x, y)
			ci := img.COffset(x, y)
			// This function may be optimized further by moving color conversion YCbCr -> RBGA out of loops
			r, g, b := color.YCbCrToRGB(img.Y[yi], img.Cb[ci], img.Cr[ci])
			avgColorsInt[0] += int(r)
			avgColorsInt[1] += int(g)
			avgColorsInt[2] += int(b)
			avgColorsInt[3] += 255
		}
	}
	for i, c := range avgColorsInt {
		avgColorsFloat[i] = float64(c) / divider
	}
	return avgColorsFloat
}

func ImageToRGBA(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}
	b := img.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(rgba, rgba.Bounds(), img, b.Min, draw.Src)
	return rgba
}

// NewLowResolutionEmbedder returns an embedder that takes an image, splits it into height*width rectangle
// and calculates the average amount of each channel (3 colors + alpha) in each rectangle.
// It produces vectors of height*width*4 numbers in range [0..1]
// E.g. for an image filled with red color (#FF0000) entirely and the following parameters: height=3, width=4
// the resulting vector would be [1,0,0,0,1,0,0,0 ... 1,0,0,0], total vector length is 48
func NewLowResolutionEmbedder(width int, height int) ImageEmbedder {
	return lowResolutionEmbedder{width, height}
}
