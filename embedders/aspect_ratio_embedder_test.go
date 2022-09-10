package embedders_test

import (
	"github.com/alef-ru/imgidx/embedders"
	"image"
	"testing"
)

func img(w, h int) *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, w, h))
}

func TestAspectRatioEmbedder(t *testing.T) {
	tests := []struct {
		name    string
		image   *image.RGBA
		want    float64
		wantErr bool
	}{
		{"nil image", nil, 0, true},
		{"height=0", img(1, 0), 0, true},
		{"width=0", img(0, 1), 0, true},
		{"1:1", img(1, 1), 0, false},
		{"2:1", img(2, 1), 0.5, false},
		{"1:2", img(1, 2), -0.5, false},
		{"1:100", img(1, 100), -0.99, false},
		{"100:1", img(100, 1), 0.99, false},
		{"100:99", img(100, 99), 0.01, false},
		{"99:100", img(99, 100), -0.01, false},
		{"100:100", img(100, 100), 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := embedders.NewAspectRatioEmbedder()
			gotV, err := r.Img2Vec(tt.image)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Img2Vec() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(gotV) != r.Dims() {
				t.Fatalf("Img2Vec() returned a vector of %v elements, %v expected", len(gotV), r.Dims())
			}
			got := gotV[0]
			if !almostEqualScalars(got, tt.want, 0.00001) {
				t.Fatalf("Img2Vec() got = %v, want %v", got, tt.want)
			}
		})
	}
}
