package index_test

import (
	"fmt"
	"github.com/alef-ru/imgidx/embedders"
	"github.com/alef-ru/imgidx/index"
	"github.com/stretchr/testify/assert"
	"image"
	"image/color"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func addPokemonsToIndex(t *testing.T, idx index.Index) {
	imgDirPath := "./testdata/pokemon"
	files, err := os.ReadDir(imgDirPath)
	if err != nil {
		t.Fatalf("failed to read files in %s : %v", imgDirPath, err)
	}
	for _, file := range files {
		_, err := index.AddImageFile(idx, path.Join(imgDirPath, file.Name()), file.Name())
		if err != nil {
			t.Fatalf("failed to add image %s : %v", file.Name(), err)
		}
	}
}

func newEmbedder() embedders.ImageEmbedder {
	return embedders.Composition([]embedders.ImageEmbedder{
		embedders.NewAspectRatioEmbedder(),
		embedders.NewColorDispersionEmbedder(),
		embedders.NewLowResolutionEmbedder(8, 8),
	})
}

func newKD3Index(t *testing.T) index.Index {

	idx, err := index.NewKDTreeImageIndex(newEmbedder())
	if err != nil {
		t.Fatalf("failed to create index : %v", err)
	}
	return idx
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

// Try to find image that is converted from PNG to JPEG and compressed
func TestIndexMatch(t *testing.T) {
	haystack := newKD3Index(t)
	addPokemonsToIndex(t, haystack)
	needlePath := "testdata/compressed_abomasnow.jpg"
	expectedImg := "abomasnow.png"
	needle, err := loadImage(needlePath)
	if err != nil {
		t.Fatalf("failed to load image %v : %v", needlePath, err)
	}
	got, _, dist, err := haystack.Nearest(needle)
	if err != nil {
		t.Fatalf("Failed to find nearest image, : %v", err)
	}
	got = filepath.Base(got)
	if got != expectedImg {
		t.Fatalf("Failed to find nearest image, got %s, want %s", got, expectedImg)
	}
	if dist > 0.025 { // Since it is the same image, the difference must be low.
		t.Fatalf("Distance between image is too far : %v", dist)
	}
}

// In this test I remove the matched image and try to search for the similar image.
// I don't care about the image found. The main point, is that the distance is big.
func TestIndexNotMatch(t *testing.T) {
	haystack := newKD3Index(t)
	addPokemonsToIndex(t, haystack)
	needlePath := "testdata/compressed_abomasnow.jpg"
	expectedImg := "abomasnow.png"
	needle, err := loadImage(needlePath)
	if err != nil {
		t.Fatalf("failed to load image %v : %v", needlePath, err)
	}

	// Remove matched image to insure, that match is impossible
	removed, err := haystack.Remove(
		func(vec embedders.Vector, uri string, attrs interface{}) bool {
			return strings.HasSuffix(uri, expectedImg)
		})
	assert.NoError(t, err, "Failed to remove image %v : %v", expectedImg)
	assert.Equal(t, 1, len(removed), "Exactly one image is expected to be removed")
	_, _, dist, err := haystack.Nearest(needle)
	if err != nil {
		t.Fatalf("Failed to find nearest image, : %v", err)
	}
	if dist < 3 { // Since we removed the matched image, the difference must be quite high
		t.Fatalf("Distance between image is too close : %v", dist)
	}
}

// In this test we try to find match for the image that differs form original significantly
// (aspect ratio, colors, format etc.).
func TestIndexWeekMatch(t *testing.T) {
	haystack := newKD3Index(t)
	addPokemonsToIndex(t, haystack)
	needlePath := "testdata/distorted_abomasnow.jpg"
	expectedImg := "abomasnow.png"
	needle, err := loadImage(needlePath)
	if err != nil {
		t.Fatalf("failed to load image %v : %v", needlePath, err)
	}

	// I don't care about the distance.
	// The main point, is that the propper image found
	got, _, _, err := haystack.Nearest(needle)
	if err != nil {
		t.Fatalf("Failed to find nearest image, : %v", err)
	}
	got = filepath.Base(got)
	if got != expectedImg {
		t.Fatalf("Failed to find nearest image, got %v, want %v", got, expectedImg)
	}

}

func generateTestImages(t *testing.T) index.Index {
	e := embedders.NewAspectRatioEmbedder()
	idx, err := index.NewKDTreeImageIndex(e)
	if err != nil {
		t.Fatalf("Failed to create idx, : %v", err)
	}
	if idx == nil {
		t.Fatalf("Failed to create idx, NewKDTreeImageIndex() returned nil, nil")
	}

	seed := map[string]image.Image{
		"1:1 image":                 image.NewRGBA(image.Rect(0, 0, 100, 100)),
		"almost 1:1 vertical image": image.NewRGBA(image.Rect(0, 0, 99, 101)),
		"2:1 image":                 image.NewRGBA(image.Rect(0, 0, 200, 100)),
		"1:2 image":                 image.NewRGBA(image.Rect(0, 0, 100, 200)),
	}
	for name, value := range seed {
		_, err := idx.AddImage(value, name, nil)
		if err != nil {
			t.Fatalf("Failed to add vector '%v' to idx, : %v", name, err)
		}
	}
	return idx
}

func TestIndexRemove(t *testing.T) {
	needle := image.NewRGBA(image.Rect(0, 0, 101, 99))

	tests := []struct {
		name           string
		f              func(vec embedders.Vector, uri string, attrs interface{}) bool
		want           []string
		nearestImgWant string
		wantErr        bool
	}{
		{
			"delete nothing",
			func(vec embedders.Vector, uri string, attrs interface{}) bool { return false },
			nil,
			"1:1 image",
			false,
		}, {
			"delete square by vec",
			func(vec embedders.Vector, uri string, attrs interface{}) bool {
				return vec[0] == 0
			},
			[]string{"1:1 image"},
			"almost 1:1 vertical image",
			false,
		}, {
			"delete square by uri",
			func(vec embedders.Vector, uri string, attrs interface{}) bool { return uri == "1:1 image" },
			[]string{"1:1 image"},
			"almost 1:1 vertical image",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx := generateTestImages(t)

			got, err := idx.Remove(tt.f)
			if (err != nil) != tt.wantErr {
				t.Fatalf("kDTreeIndex.Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, got)
			nearestImgGot, _, _, err := idx.Nearest(needle)
			if err != nil {
				t.Fatalf("Failed to find nearest image, : %v", err)
			}
			if tt.nearestImgWant != nearestImgGot {
				t.Fatalf("Failed to find nearest image, got '%v', want '%v'", nearestImgGot, tt.nearestImgWant)
			}
		})
	}
}

func TestIndexConcurrentWrite(t *testing.T) {
	const iterations = 100
	deletionResults := make(chan []string, iterations)
	extraImage := image.NewRGBA(image.Rect(0, 0, 100, 100))
	removeExtraImages := func(vec embedders.Vector, uri string, attrs interface{}) bool {
		return attrs == "extra"
	}
	idx := newKD3Index(t)
	addPokemonsToIndex(t, idx)
	originalIdxLen := idx.GetCount()
	var wg sync.WaitGroup
	wg.Add(iterations * 2)
	for i := 0; i < iterations; i++ {
		go func(i int) {
			defer wg.Done()
			uri := fmt.Sprintf("files://./image_%d", i)
			_, err := idx.AddImage(extraImage, uri, "extra")
			if err != nil {
				panic("Failed to add extra image to index")
			}
		}(i)
	}
	for i := 0; i < iterations; i++ {
		go func(i int) {
			defer wg.Done()
			deleted, err := idx.Remove(removeExtraImages)
			if err != nil {
				panic("Failed to remove extra images from index")
			}
			deletionResults <- deleted
		}(i)
	}
	wg.Wait()
	close(deletionResults)
	removed, err := idx.Remove(removeExtraImages)
	deletedTotal := len(removed)
	if err != nil {
		t.Fatalf("Failed to remove extra images from index, : %v", err)
	}
	for removed := range deletionResults {
		deletedTotal += len(removed)
	}
	if deletedTotal != iterations {
		t.Fatalf("%v images was expected to be removed, %v was removed in fact", iterations, deletedTotal)
	}
	if idx.GetCount() != originalIdxLen {
		t.Fatalf("%v images was expected to be in index, %v in fact", originalIdxLen, idx.GetCount())
	}
}

func TestAddImageUrl(t *testing.T) {
	server := runTestImgHttpServer()
	defer server.Close()

	imgUrl := server.URL + "/abra.png"
	imgPath := "./testdata/pokemon/abra.png"
	idx := newKD3Index(t)
	urlVec, err := index.AddImageUrl(idx, imgUrl, "form url")
	if err != nil {
		t.Fatalf("Failed to add image from url %v : %v", imgUrl, err)
	}

	fileVec, err := index.AddImageFile(idx, imgPath, "form file")
	if err != nil {
		t.Fatalf("Failed to add image from file %v : %v", imgPath, err)
	}

	if !reflect.DeepEqual(urlVec, fileVec) {
		t.Fatalf("Vectors from url and file are not equal")
	}
}

func runTestImgHttpServer() *httptest.Server {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			imgPath := path.Join("testdata/pokemon/", r.URL.Path[1:])
			http.ServeFile(w, r, imgPath)
		}))
	return server
}

func TestUniqueUris(t *testing.T) {
	testImgPath := "testdata/pokemon/abra.png"
	idx := newKD3Index(t)
	addPokemonsToIndex(t, idx)
	cnt := idx.GetCount()

	vec, err := index.AddImageFile(idx, testImgPath, "already exists")
	assert.NotNil(t, err, "Error expected")
	assert.Nil(t, vec, "The result was expected to be nil")
	assert.Equal(t, cnt, idx.GetCount(), "The index size was expected to remain the same")

	removed, err := idx.Remove(func(vec embedders.Vector, uri string, attrs interface{}) bool {
		return strings.HasSuffix(uri, testImgPath)
	})
	assert.Nil(t, err, "Failed to delete image %v", testImgPath)
	assert.Equal(t, cnt-1, idx.GetCount(), "The index size was expected to decrease")
	assert.Equal(t, len(removed), 1, "Exactly one image was expected to be removed")

	vec, err = index.AddImageFile(idx, "testdata/pokemon/abra.png", "already exists")
	assert.Nil(t, err, "Error was not expected")
	assert.NotNil(t, vec, "The result was not expected to be nil")
	assert.Equal(t, cnt, idx.GetCount(), "The index size was expected to remain the same")

}
