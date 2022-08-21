package index_test

import (
	"github.com/alef-ru/imgidx/embedders"
	"github.com/alef-ru/imgidx/index"
	"image"
	"testing"
)

func Test_kDTreeMatcher_Happy_Pass(t *testing.T) {
	idx := testData(t)
	needle := image.NewRGBA(image.Rect(0, 0, 101, 99))
	got, dist, err := idx.Nearest(needle)
	if err != nil {
		t.Fatalf("Failed to find nearest image, : %v", err)
	}
	if got != "1:1 image" {
		t.Fatalf("Failed to find nearest image, got '%v', want '1:1 image'", got)
	}
	if dist == 0 {
		t.Fatalf("Distance must be >0, because images are not the same")
	}
}

func img(w, h int) image.Image {
	return image.NewRGBA(image.Rect(0, 0, w, h))
}

func testData(t *testing.T) index.Index {
	var e embedders.ImageEmbedder = embedders.AspectRatioEmbedder{}
	idx, err := index.NewKDTreeImageIndex(e)
	if err != nil {
		t.Fatalf("Failed to create idx, : %v", err)
	}
	if idx == nil {
		t.Fatalf("Failed to create idx, NewKDTreeImageIndex() returned nil, nil")
	}
	seed := map[string]image.Image{
		"1:1 image":                 img(100, 100),
		"almost 1:1 vertical image": img(99, 101),
		"2:1 image":                 img(200, 100),
		"1:2 image":                 img(100, 200),
	}
	for name, value := range seed {
		_, err := idx.AddImage(value, name)
		if err != nil {
			t.Fatalf("Failed to add vector '%v' to idx, : %v", name, err)
		}
	}
	return idx
}

func Test_kDTreeIndex_Remove(t *testing.T) {
	needle := image.NewRGBA(image.Rect(0, 0, 101, 99))

	tests := []struct {
		name           string
		f              func(vec embedders.Vector, attrs interface{}) bool
		want           int
		nearestImgWant string
		wantErr        bool
	}{
		{
			"delete nothing",
			func(vec embedders.Vector, attrs interface{}) bool { return false },
			0,
			"1:1 image",
			false,
		}, {
			"delete square by vec",
			func(vec embedders.Vector, attrs interface{}) bool {
				return vec[0] == 0
			},
			1,
			"almost 1:1 vertical image",
			false,
		}, {
			"delete square by attrs",
			func(vec embedders.Vector, attrs interface{}) bool { return attrs == "1:1 image" },
			1,
			"almost 1:1 vertical image",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx := testData(t)

			got, err := idx.Remove(tt.f)
			if (err != nil) != tt.wantErr {
				t.Fatalf("kDTreeIndex.Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("kDTreeIndex.Remove() = %v, want %v", got, tt.want)
			}
			nearestImgGot, _, err := idx.Nearest(needle)
			if err != nil {
				t.Fatalf("Failed to find nearest image, : %v", err)
			}
			if tt.nearestImgWant != nearestImgGot {
				t.Fatalf("Failed to find nearest image, got '%v', want '%v'", nearestImgGot, tt.nearestImgWant)
			}
		})
	}
}
