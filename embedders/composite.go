package embedders

import "image"

type compositeEmbedder struct {
	Embedders []ImageEmbedder
}

// Composition returns a new embedder that converts image into a vector produced as a concatenation of the given embedders' vectors
func Composition(embedders []ImageEmbedder) ImageEmbedder {
	return compositeEmbedder{Embedders: embedders}
}

func (a compositeEmbedder) Img2Vec(image *image.RGBA) (Vector, error) {
	if image == nil {
		return nil, ErrEmptyImage
	}
	var v Vector
	for _, e := range a.Embedders {
		vec, err := e.Img2Vec(image)
		if err != nil {
			return nil, err
		}
		v = append(v, vec...)
	}
	return v, nil
}

func (a compositeEmbedder) Dims() int {
	var dims int
	for _, e := range a.Embedders {
		dims += e.Dims()
	}
	return dims
}
