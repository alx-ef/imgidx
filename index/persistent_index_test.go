package index_test

import (
	"github.com/alef-ru/imgidx/embedders"
	"github.com/alef-ru/imgidx/index"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"os"
	"strings"
	"testing"
)

func TestPersistentIndexHappyPass(t *testing.T) {
	const pathToDB = "./tmp_imgidx.db"
	const testImgPath = "testdata/pokemon/absol.png"
	t.Cleanup(func() {
		_ = os.Remove(pathToDB)
	})

	// Create a new index from scratch. It is supposed to be empty
	dialector := sqlite.Open(pathToDB)
	kd3idx := newKD3Index(t)
	var idx index.Index
	idx, err := index.NewPersistentIndex(dialector, kd3idx)
	assert.NoError(t, err, "Failed to create PersistentIndex from scratch")
	assert.Equal(t, 0, idx.GetCount())

	// Add some images to the index
	addPokemonsToIndex(t, idx)
	cnt := idx.GetCount()
	assert.NotEqual(t, 0, cnt)

	// Create a new index with the same parameters. It is supposed to load all the data from the DB
	dialector = sqlite.Open(pathToDB)
	kd3idx = newKD3Index(t)
	idx, err = index.NewPersistentIndex(dialector, kd3idx)
	assert.NoError(t, err, "Failed to create PersistentIndex with existing SQLite file")
	assert.Equal(t, cnt, idx.GetCount())
	assert.Equal(t, cnt, kd3idx.GetCount())
	vec, err := index.AddImageFile(idx, testImgPath, nil)
	assert.Error(t, err)
	assert.Nil(t, vec)

	// Remove an image from the index
	removed, err := idx.Remove(func(vec embedders.Vector, uri string, attrs interface{}) bool {
		return strings.HasSuffix(uri, testImgPath)
	})
	assert.NoError(t, err, "Failed to remove image from index")
	assert.Equal(t, 1, len(removed))
	assert.Equal(t, cnt-1, idx.GetCount())

	// Reload the index and check that the image was removed again
	dialector = sqlite.Open(pathToDB)
	kd3idx = newKD3Index(t)
	idx, err = index.NewPersistentIndex(dialector, kd3idx)
	assert.NoError(t, err, "Failed to create PersistentIndex with existing SQLite file")
	assert.Equal(t, cnt-1, idx.GetCount())
	testImg, err := loadImage(testImgPath)
	assert.NoError(t, err, "Failed to load image %s", testImgPath)
	_, _, dist, err := idx.Nearest(testImg)
	assert.NoError(t, err, "Failed to find nearest image")
	assert.True(t, dist > 0.1, "Nearest image is too close, dist: %f. Seems like it is the same image, "+
		"even though it was supposed to be removed", dist)
}

func TestNewPersistentIndexInvalidDB(t *testing.T) {
	_, err := index.NewPersistentIndex(sqlite.Open("/invalid/path/to/db"), newKD3Index(t))
	assert.ErrorContains(t, err, "no such file or directory")
}

func TestPersistentIndexDBWriteFailure(t *testing.T) {
	const pathToDB = "tmp_err.db"
	const testImgPath = "testdata/pokemon/absol.png"
	idx, err := index.NewPersistentIndex(sqlite.Open(pathToDB), newKD3Index(t))
	assert.NoError(t, os.Remove(pathToDB))
	_, err = index.AddImageFile(idx, testImgPath, nil)
	assert.ErrorContains(t, err, "failed to save image embed to DB")
}

func TestPersistentIndexMigrationFailure(t *testing.T) {
	_, err := index.NewPersistentIndex(sqlite.Open("testdata/invalid.db"), newKD3Index(t))
	assert.ErrorContains(t, err, "failed to migrate db")
}

func makeTestPersistentIndex(t *testing.T) *index.PersistentIndex {
	const pathToDB = "tmp_test.db"
	t.Cleanup(func() {
		_ = os.Remove(pathToDB)
	})
	idx, err := index.NewPersistentIndex(sqlite.Open(pathToDB), newKD3Index(t))
	assert.NoError(t, err, "Failed to create PersistentIndex")
	return idx
}

func TestPersistentIndexAddVector(t *testing.T) {
	idx := makeTestPersistentIndex(t)
	vec := make(embedders.Vector, newEmbedder().Dims())
	assert.NoError(t, idx.AddVector(vec, "test.png", nil))
	assert.ErrorContains(t, idx.AddVector(vec, "test.png", nil),
		"failed to save image embed to DB: UNIQUE constraint failed: img_embeds.uri")
}

func TestPersistentIndexRemoveNothing(t *testing.T) {
	idx := makeTestPersistentIndex(t)
	addPokemonsToIndex(t, idx)
	cnt := idx.GetCount()
	assert.NotEqual(t, 0, cnt)
	removed, err := idx.Remove(func(vec embedders.Vector, uri string, attrs interface{}) bool { return false })
	assert.NoError(t, err, "Failed to remove image from index")
	assert.Equal(t, 0, len(removed))
	assert.Equal(t, cnt, idx.GetCount())
}
