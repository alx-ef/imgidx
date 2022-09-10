package embedders_test

import (
	"image"
	_ "image/jpeg"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alef-ru/imgidx/embedders"
)

// Before optimisation:
// BenchmarkLowResEmbedder_Img2Vec_NoConversion_8_8-10         9410            124503 ns/op           42048 B/op      10001 allocs/op
//
// After optimisation:
// BenchmarkLowResEmbedder_Img2Vec_NoConversion_8_8-10        23792             48998 ns/op            2048 B/op          1 allocs/op
func BenchmarkLowResEmbedder_Img2Vec_NoConversion_8_8(b *testing.B) {
	e := embedders.NewLowResolutionEmbedder(8, 8)
	img := createTestImage(100, 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := e.Img2Vec(img)
		assert.NoError(b, err)
	}
}

// Before optimisation:
// BenchmarkLowResEmbedder_Img2Vec_NoConversion_8_8-10          264           4414829 ns/op          840916 B/op     262145 allocs/op
//
// After optimisation:
// BenchmarkLowResEmbedder_Img2Vec_NoConversion_8_8-10           22          50949258 ns/op        67115143 B/op        130 allocs/op
// BenchmarkLowResEmbedder_Img2Vec_NoConversion_8_8-10          582           2018670 ns/op         1050691 B/op          3 allocs/op
// BenchmarkLowResEmbedder_Img2Vec_NoConversion_8_8-10          560           2007634 ns/op         1050689 B/op          3 allocs/op
// BenchmarkLowResEmbedder_Img2Vec_NoConversion_8_8-10          189           6124052 ns/op         2099218 B/op     524289 allocs/op
// BenchmarkLowResEmbedder_Img2Vec_NoConversion_8_8-10        24303             48790 ns/op            2048 B/op           1 allocs/op
func BenchmarkLowResEmbedder_Img2Vec_Jpeg_8_8(b *testing.B) {
	e := embedders.NewLowResolutionEmbedder(8, 8)
	path := "testdata/lena.jpeg"
	img, err := loadImage(path)
	assert.NoError(b, err, "Failed to load test image %s", path)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := e.Img2Vec(img)
		assert.NoError(b, err)
	}
}

// Before optimisation:
// BenchmarkDispersionEmbedder_Img2Vec_NoConversion_8_8-10    	    6752	    175848 ns/op	   40056 B/op	   10002 allocs/op
//
// After optimisation:
// BenchmarkDispersionEmbedder_Img2Vec_NoConversion_8_8-10    	   18382	     65366 ns/op	      56 B/op	       2 allocs/op
func BenchmarkDispersionEmbedder_Img2Vec_NoConversion_8_8(b *testing.B) {
	e := embedders.NewColorDispersionEmbedder()
	img := createTestImage(100, 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := e.Img2Vec(img)
		assert.NoError(b, err)
	}
}

func TestRGBAvsYCbCr(t *testing.T) {
	e := embedders.Composition([]embedders.ImageEmbedder{
		embedders.NewLowResolutionEmbedder(8, 8),
		embedders.NewColorDispersionEmbedder(),
	})
	path := "testdata/lena.jpeg"
	ycbcrImg, err := loadImage(path)
	assert.NoErrorf(t, err, "Failed to load test image %s", path)
	assert.IsType(t, &image.YCbCr{}, ycbcrImg)
	rgbaImg := embedders.ImageToRGBA(ycbcrImg)
	assert.IsType(t, &image.RGBA{}, rgbaImg)

	v1, err := e.Img2Vec(rgbaImg)
	assert.NoError(t, err)
	v2, err := e.Img2Vec(ycbcrImg)
	assert.NoError(t, err)
	assert.Equal(t, v1, v2)
}
