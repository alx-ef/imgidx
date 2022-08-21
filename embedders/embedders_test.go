package embedders_test

import (
	"image"
	"image/color"
	_ "image/png" // register PNG decoder
	"math"
	"math/rand"
	"os"
)

func almostEqualScalars(a, b float64, threshold float64) bool {
	return math.Abs(a-b) <= threshold
}

func almostEqualSlices(a, b []float64, threshold float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if math.Abs(a[i]-b[i]) > threshold {
			return false
		}
	}
	return true
}

func loadImage(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	img.ColorModel().Convert(color.RGBA{})
	return img, err
}

func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			switch {
			case x < width/2 && y < height/2: // upper left quadrant
				img.Set(x, y, color.White)
			case x >= width/2 && y < height/2: // upper right quadrant
				img.Set(x, y, color.Black)
			case x < width/2 && y >= height/2: // lower left quadrant
				img.Set(x, y, color.Color(color.RGBA{255, 0, 0, 255})) // pure red
			case x >= width/2 && y >= height/2: // lower right quadrant
				img.Set(x, y, color.Color(color.RGBA{0, 127, 0, 127})) // 50% transparent green
			}
		}
	}
	//saveAsPng(img, "testdata/tst.png")
	return img
}

func createMonochromeImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.Color(color.RGBA{100, 0, 200, 255})) // pure red
		}
	}
	return img
}

func createNoisyImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			rgba := make([]byte, 4)
			rand.Read(rgba)
			img.Set(x, y, color.Color(color.RGBA{rgba[0], rgba[1], rgba[2], rgba[3]})) // random color
		}
	}
	return img
}

func createMaxDispersionImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if (x+y)%2 == 0 { //odds white and evens black
				img.Set(x, y, color.Color(color.White))
			} else {
				img.Set(x, y, color.Color(color.Black))
			}
		}
	}
	return img
}
