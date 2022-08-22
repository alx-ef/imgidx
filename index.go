package imgidx

import (
	"fmt"
	"github.com/alef-ru/imgidx/embedders"
	"image"
	_ "image/gif"  // register GIF decoder
	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder
	"os"
	"sync"
)
import "gonum.org/v1/gonum/spatial/kdtree"

// Index is an index of images that be searched for nearest neighbors: most similar images
// It's not supposed to keep the images in memory, but only some compact representation of the images (the vectors ).
type Index interface {
	// AddImage adds the image img to the index and returns its vector representation.
	//
	// attrs is supposed to contain image's attributes, such as URL, UUID, id, type, etc.
	// It's stored by the index as is and returned by the Nearest method.
	//
	// The Vector is supposed to be stored in persistent storage, so the Index state is possible to restore
	// without reindexing all the images.
	AddImage(img image.Image, attrs interface{}) (embedders.Vector, error)

	// AddVector adds a pre-calculated vector to the index.
	// This method is supposed to be used when restoring the index from a persistent storage.
	//
	// The vec is supposed to be a Vector returned by Index.AddImage().
	//
	// attrs is supposed to contain image's attributes, such as URL, UUID, id, type, etc.
	// It's stored by the index as-is and returned by the Nearest method.
	//
	// Vector length must match the index's number of dimensions
	AddVector(vec embedders.Vector, attrs interface{}) error

	// Nearest embeds the image img into a vector searches for the nearest neighbor in the index.
	// It returns the found image's attributes as-is and the distance between the given and found images.
	// Image will always be found unless the index is empty, regardless on the distances.
	// It's up to caller to consider it as "match" or "not found" depending on the distance between images.
	Nearest(img image.Image) (attrs interface{}, distance float64, err error)

	// Remove checks each image representation with the passed function,
	// and removes the image if the function returns true.
	//
	// Remove returns the number removed of images ( >=0 )
	//
	// This function is implemented this way, not like Remove(id) for two reasons:
	// - Images in the index have no unique identifier. It's possible to have one in attrs, but the index doesn't work with it.
	// - The underlining kdtree implementation doesn't support removal, so we have to rebuild the index, and it's better to do it in one go.
	Remove(func(vec embedders.Vector, attrs interface{}) bool) (int, error)

	// GetCount returns the number of images in the index.
	GetCount() int
}

type kDTreeIndex struct {
	tree     *kdtree.Tree
	embedder embedders.ImageEmbedder
	dims     int
	lock     sync.RWMutex
}

func (idx *kDTreeIndex) AddVector(vec embedders.Vector, attrs interface{}) error {
	if len(vec) != idx.dims {
		return fmt.Errorf("vector has %d dimensions. Expected %d", len(vec), idx.dims)
	}
	idx.lock.Lock()
	defer idx.lock.Unlock()
	idx.tree.Insert(embed{kdtree.Point(vec), attrs}, false)
	return nil
}

func (idx *kDTreeIndex) Nearest(img image.Image) (attrs interface{}, distance float64, err error) {
	vec, err := idx.embedder.Img2Vec(img)
	if err != nil {
		return nil, 0, err
	}
	idx.lock.RLock()
	defer idx.lock.RUnlock()
	got, dist := idx.tree.Nearest(embed{kdtree.Point(vec), nil})
	embd, ok := got.(embed)
	if !ok {
		return nil, 0, fmt.Errorf("got %T, expected embed", got)
	}
	return embd.attrs, dist, nil
}

func (idx *kDTreeIndex) Remove(f func(vec embedders.Vector, attrs interface{}) bool) (int, error) {
	//FixMe: it seems inefficient to rebuild the index every time, but it's the easiest way to implement Remove
	keep := make(embeds, 0)
	var removeCnt int
	var err error
	idx.lock.Lock()
	defer idx.lock.Unlock()
	idx.tree.Do(func(c kdtree.Comparable, _ *kdtree.Bounding, _ int) bool {
		embd, ok := c.(embed)
		if !ok {
			err = fmt.Errorf("KDTree contained %T, expected embed only", c)
			return true
		}
		if f(embedders.Vector(embd.Point), embd.attrs) {
			removeCnt += 1
		} else {
			keep = append(keep, embd)
		}
		return false
	})
	if err != nil {
		return 0, err
	}
	idx.tree = kdtree.New(keep, false)
	return removeCnt, nil
}

func (idx *kDTreeIndex) AddImage(img image.Image, attrs interface{}) (embedders.Vector, error) {
	vec, err := idx.embedder.Img2Vec(img)
	if err != nil {
		return nil, err
	}
	err = idx.AddVector(vec, attrs)
	if err != nil {
		return nil, err
	}
	return vec, nil
}

func (idx *kDTreeIndex) GetCount() int {
	idx.lock.RLock() //FixMe: I'm not if lock is needed here
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
	return idx.AddImage(img, attrs)
}
