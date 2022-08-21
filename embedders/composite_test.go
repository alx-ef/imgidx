package embedders_test

import (
	"github.com/alef-ru/imgidx/embedders"
	"testing"
)

func TestCompositeEmbedder(t *testing.T) {
	ce := embedders.Composition([]embedders.ImageEmbedder{
		embedders.NewAspectRatioEmbedder(),
		embedders.NewColorDispersionEmbedder(),
		embedders.NewLowResolutionEmbedder(8, 8),
	})

	dimsWant := 1 + 3 + 8*8*4
	dimsGot := ce.Dims()
	if dimsGot != dimsWant {
		t.Errorf("Dims() = %v, want %v", dimsGot, dimsWant)
	}
	vec, err := ce.Img2Vec(createMaxDispersionImage(200, 100))
	if err != nil {
		t.Errorf("Img2Vec() returned erroor %v", err)
	}
	if len(vec) != dimsWant {
		t.Errorf("len(Img2Vec()) = %v, want %v", len(vec), dimsWant)
	}
	if !almostEqualScalars(vec[0], 0.5, 0.00001) {
		t.Errorf("vec[0], produced by aspectRatioEmbedder = %v, want 0.5", vec[0])
	}
	disp := vec[1] * vec[2] * vec[3]
	if 1 < disp || disp < 0.9 {
		t.Errorf("vector components produced by colorDispersionEmbedder have unexpected value")
	}

}
