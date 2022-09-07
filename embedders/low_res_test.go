package embedders_test

import (
	"image"
	"math"
	"testing"

	"github.com/alef-ru/imgidx/embedders"
)

func TestLowResEmbedderImg2Vec(t *testing.T) {
	type args struct {
		img    image.Image
		height int
		width  int
	}
	tests := []struct {
		name    string
		args    args
		want    []float64
		wantErr bool
	}{
		{
			"2*2 test image",
			args{
				createTestImage(100, 100),
				2,
				2,
			},
			[]float64{
				1, 1, 1, 1, // upper left quadrant- white
				0, 0, 0, 1, // upper right quadrant - black
				1, 0, 0, 1, // lower left quadrant - red
				0, 0.5, 0, 0.5, // lower right quadrant - 50% transparent 50% green
			},
			false,
		}, {
			"4*2 test image",
			args{
				createTestImage(100, 100),
				4,
				2,
			},
			[]float64{
				1, 1, 1, 1, 1, 1, 1, 1, // upper left quadrant- white
				0, 0, 0, 1, 0, 0, 0, 1, // upper right quadrant - black
				1, 0, 0, 1, 1, 0, 0, 1, // lower left quadrant - red
				0, 0.5, 0, 0.5, 0, 0.5, 0, 0.5, // lower right quadrant - 50% transparent 50% green
			},
			false,
		}, {
			"3*2 test image",
			args{
				createTestImage(100, 200),
				3,
				2,
			},
			[]float64{
				1, 1, 1, 1, // upper left quadrant- white
				0.5, 0.5, 0.5, 1, // upper medium quadrant- black and white
				0, 0, 0, 1, // upper right quadrant - black
				1, 0, 0, 1, // lower left quadrant - red
				0.5, 0.25, 0, 0.75, // lower medium quadrant - red / 50% transparent 50% green
				0, 0.5, 0, 0.5, // lower right quadrant - 50% transparent 50% green
			},
			false,
		}, {
			"image too small",
			args{
				createTestImage(2, 2),
				3,
				2,
			},
			nil,
			true,
		},
		{
			"wrong embedder parameters",
			args{
				createTestImage(100, 100),
				1,
				0,
			},
			nil,
			true,
		}, {
			"0x0 image",
			args{
				createTestImage(0, 0),
				2,
				2,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := embedders.NewLowResolutionEmbedder(tt.args.height, tt.args.width)
			got, err := e.Img2Vec(tt.args.img)
			for i, v := range got {
				got[i] = math.Round(v*1000) / 1000
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("image2Vec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !almostEqualSlices(got, tt.want, 0.02) {
				t.Errorf("image2Vec() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLowResEmbedderDims(t *testing.T) {
	e := embedders.NewLowResolutionEmbedder(10, 20)
	img := createTestImage(100, 100)
	size := e.Dims()
	vec, err := e.Img2Vec(img)
	if err != nil {
		t.Errorf("Img2Vec() returned error: %v", err)
	}
	if len(vec) != size {
		t.Errorf("GetSize() returned %v, but the actual vector size is got: %v", size, len(vec))
	}
}

// Before optimisation:
// BenchmarkLowResEmbedder_Img2Vec_8_8-10              9410            124503 ns/op           42048 B/op      10001 allocs/op
//
// After optimisation:
// BenchmarkLowResEmbedder_Img2Vec_8_8-10             23830             49167 ns/op            2048 B/op          1 allocs/op
func BenchmarkLowResEmbedder_Img2Vec_8_8(b *testing.B) {
	e := embedders.NewLowResolutionEmbedder(8, 8)
	img := createTestImage(100, 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Img2Vec(img)
	}
}
