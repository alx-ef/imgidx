package imgidx

import (
	_ "image/jpeg"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// goos: darwin, goarch: arm64
// Before optimisation
// BenchmarkComposite_Img2Vec_Jpeg-10           134           8808894 ns/op         1424701 B/op     443733 allocs/op
// After optimisation
// BenchmarkComposite_Img2Vec_Jpeg-10           480           2379291 ns/op          602853 B/op         11 allocs/op
// ok      github.com/alef-ru/imgidx       2.151s
func BenchmarkComposite_Img2Vec_Jpeg(b *testing.B) {
	idx, err := NewCompositeIndex(8, 8)
	assert.NoError(b, err, "Failed to create index")
	path := "testdata/distorted_abomasnow.jpg"
	img, err := readImageFile(path)
	assert.NoErrorf(b, err, "Failed to read image %s", path)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := idx.AddImage(img, strconv.Itoa(i), nil)
		assert.NoError(b, err)
	}
}
