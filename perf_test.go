package imgidx

import (
	_ "image/jpeg"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
