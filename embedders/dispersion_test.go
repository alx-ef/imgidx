package embedders_test

import (
	"github.com/alef-ru/imgidx/embedders"
	"image"
	"testing"
)

func TestColorDispersionEmbedder_Img2Vec(t *testing.T) {
	const width, height = 201, 99
	embedder := embedders.NewColorDispersionEmbedder()
	path := "testdata/lenna.png"
	lennaPng, err := loadImage(path)
	if err != nil {
		t.Errorf("failed to read file %v: %v", path, err)
	}
	tests := []struct {
		name     string
		image    *image.RGBA
		wantFrom float64
		wantTo   float64
		wantErr  bool
	}{
		{
			"Monochrome Image",
			createMonochromeImage(width, height),
			0,
			0.0001,
			false,
		}, {
			"Noisy Image",
			createNoisyImage(width, height),
			0.4,
			0.6,
			false,
		}, {
			"lenna.png",
			lennaPng,
			0.1,
			0.5,
			false,
		}, {
			"Max Dispersion Image",
			createMaxDispersionImage(width, height),
			0.99,
			1.00,
			false,
		}, {
			name:    "nil Image",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := embedder.Img2Vec(tt.image)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Error was expected")
				}
				if got != nil {
					t.Fatalf("Result was expected to be nil")
				}
				return
			}
			if err != nil {
				t.Errorf("dispersionImage2Vec() should not return an error")
			}
			if got == nil {
				t.Errorf("dispersionImage2Vec() should return a vector")
			}
			if len(got) != embedder.Dims() {
				t.Fatalf("Img2Vec() returned a vector of %v elements, %v expected", len(got), embedder.Dims())
			}
			for _, f := range got {
				if f > tt.wantTo || f < tt.wantFrom {
					t.Errorf("Img2Vec() expected to returned numbers in range %v .. %v, it returned %v",
						tt.wantFrom, tt.wantTo, got)
				}
			}
		})
	}
}
