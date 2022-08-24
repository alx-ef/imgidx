package index

import (
	"fmt"
	"github.com/alef-ru/imgidx/embedders"
	"image"
	_ "image/gif"  // register GIF decoder
	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)
import "gonum.org/v1/gonum/spatial/kdtree"

// Index is an index of images that be searched for nearest neighbors: most similar images
// It's not supposed to keep the images in memory, but only some compact representation of the images (the vectors ).
type Index interface {
	// AddImage adds the image img to the index and returns its vector representation.
	//
	// URI is the URI of the image: "https://...", "files://.." or anything else. It's supposed to be unique.
	// If the image with the URI is already in the index, it causes an error.
	//
	// Attributes is supposed to contain image's attributes, such as URL, UUID, id, type, etc.
	// It's stored by the index as is and returned by the Nearest method.
	//
	// The Vector is supposed to be stored in persistent storage, so the Index state is possible to restore
	// without reindexing all the images.
	AddImage(img image.Image, uri string, attrs interface{}) (embedders.Vector, error)

	// AddVector adds a pre-calculated vector to the index.
	// This method is supposed to be used when restoring the index from a persistent storage.
	//
	// The vec is supposed to be a Vector returned by Index.AddImage().
	//
	// The URI and Attributes are the same as in AddImage().
	//
	// Vector length must match the index's number of dimensions
	AddVector(vec embedders.Vector, uri string, attrs interface{}) error

	// Nearest embeds the image img into a vector searches for the nearest neighbor in the index.
	// It returns the found image's attributes as-is and the distance between the given and found images.
	// Image will always be found unless the index is empty, regardless on the distances.
	// It's up to caller to consider it as "match" or "not found" depending on the distance between images.
	Nearest(img image.Image) (uri string, attrs interface{}, distance float64, err error)

	// Remove checks each image representation with the passed function,
	// and removes the image if the function returns true.
	//
	// Remove returns URIs of removed of images.
	//
	// This function is implemented this way, not like Remove(id) for two reasons:
	// - Images in the index have no unique identifier. It's possible to have one in Attributes, but the index doesn't work with it.
	// - The underlining kdtree implementation doesn't support removal, so we have to rebuild the index, and it's better to do it in one go.
	Remove(func(vec embedders.Vector, uri string, attrs interface{}) bool) (removed []string, err error)

	// GetCount returns the number of images in the index.
	GetCount() int
}

type kDTreeIndex struct {
	tree     *kdtree.Tree
	embedder embedders.ImageEmbedder
	dims     int
	lock     sync.RWMutex
	uris     map[string]bool
}

func (idx *kDTreeIndex) AddVector(vec embedders.Vector, uri string, attrs interface{}) error {
	if len(vec) != idx.dims {
		return fmt.Errorf("vector has %d dimensions. Expected %d", len(vec), idx.dims)
	}
	idx.lock.Lock()
	defer idx.lock.Unlock()
	if idx.uris[uri] {
		return fmt.Errorf("image with URI %s is already in the index", uri)
	}
	idx.tree.Insert(ImgEmbed{URI: uri, Vector: kdtree.Point(vec), Attributes: attrs}, false)
	idx.uris[uri] = true
	return nil
}

func (idx *kDTreeIndex) Nearest(img image.Image) (uri string, attrs interface{}, distance float64, err error) {
	vec, err := idx.embedder.Img2Vec(img)
	if err != nil {
		return "", nil, 0, err
	}
	idx.lock.RLock()
	defer idx.lock.RUnlock()
	got, dist := idx.tree.Nearest(ImgEmbed{Vector: kdtree.Point(vec)})
	embd, ok := got.(ImgEmbed)
	if !ok {
		return "", nil, 0, fmt.Errorf("got %T, expected ImgEmbed", got)
	}
	return embd.URI, embd.Attributes, dist, nil
}

func (idx *kDTreeIndex) Remove(f func(vec embedders.Vector, uri string, attrs interface{}) bool) ([]string, error) {
	//FixMe: it seems inefficient to rebuild the index every time, but it's the easiest way to implement Remove
	keep := make(embeds, 0)
	var remove []string
	var err error
	idx.lock.Lock()
	defer idx.lock.Unlock()
	idx.tree.Do(func(c kdtree.Comparable, _ *kdtree.Bounding, _ int) bool {
		embd, ok := c.(ImgEmbed)
		if !ok {
			err = fmt.Errorf("KDTree contained %T, expected ImgEmbed only", c)
			return true
		}
		if f(embedders.Vector(embd.Vector), embd.URI, embd.Attributes) {
			remove = append(remove, embd.URI)
		} else {
			keep = append(keep, embd)
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	if len(remove) != 0 {
		idx.tree = kdtree.New(keep, false)
		idx.uris = make(map[string]bool, len(keep))
		for _, embd := range keep {
			idx.uris[embd.URI] = true
		}
	}
	return remove, nil
}

func (idx *kDTreeIndex) AddImage(img image.Image, uri string, attrs interface{}) (embedders.Vector, error) {
	//log.Println("Adding image", URI)
	vec, err := idx.embedder.Img2Vec(img)
	if err != nil {
		return nil, err
	}
	err = idx.AddVector(vec, uri, attrs)
	if err != nil {
		return nil, err
	}
	return vec, nil
}

func (idx *kDTreeIndex) GetCount() int {
	idx.lock.RLock()
	defer idx.lock.RUnlock()
	return idx.tree.Count
}

func NewKDTreeImageIndex(embedder embedders.ImageEmbedder) (Index, error) {
	var index kDTreeIndex
	if embedder == nil {
		return nil, fmt.Errorf("embedder is nil")
	}
	index.dims = embedder.Dims()
	if index.dims <= 0 {
		return nil, fmt.Errorf("embedder has %d dimensions. A positive number expected", index.dims)
	}
	index.embedder = embedder
	index.tree = kdtree.New(make(embeds, 0), false)
	index.uris = make(map[string]bool)
	return &index, nil
}

func AddImageFile(idx Index, path string, attrs interface{}) (vec embedders.Vector, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file %s: %w", path, err)
	}
	defer func() {
		closingErr := f.Close()
		if err == nil {
			err = closingErr
		}
	}()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}
	uri := "file://" + filepath.Join(wd, path)
	return idx.AddImage(img, uri, attrs)
}

func AddImageUrl(idx Index, url string, attrs interface{}) (embedders.Vector, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		closingErr := Body.Close()
		if err != nil {
			err = closingErr
		}
	}(res.Body)
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("received code %d, 200 expected", res.StatusCode)
	}
	img, _, err := image.Decode(res.Body)
	return idx.AddImage(img, url, attrs)
}
