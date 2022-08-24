package index

import (
	"fmt"
	"github.com/alef-ru/imgidx/embedders"
	"gonum.org/v1/gonum/spatial/kdtree"
	"gorm.io/gorm"
	"image"
	"sync"
)

type PersistentIndex struct {
	db    *gorm.DB
	inIdx Index
	lock  sync.Mutex
}

func (idx *PersistentIndex) saveVec(vec embedders.Vector, uri string, attrs interface{}) error {
	idx.lock.Lock()
	defer idx.lock.Unlock()
	embed := ImgEmbed{
		URI:        uri,
		Vector:     kdtree.Point(vec),
		Attributes: attrs,
	}
	result := idx.db.Create(&embed)
	if result.Error != nil {
		return fmt.Errorf("failed to save image embed to DB: %w", result.Error)
	}
	return nil
}

func (idx *PersistentIndex) AddImage(img image.Image, uri string, attrs interface{}) (embedders.Vector, error) {
	vec, err := idx.inIdx.AddImage(img, uri, attrs)
	if err != nil {
		return nil, err
	}
	err = idx.saveVec(vec, uri, attrs)
	if err != nil {
		return nil, err
	}
	return vec, nil
}

func (idx *PersistentIndex) AddVector(vec embedders.Vector, uri string, attrs interface{}) error {
	if err := idx.saveVec(vec, uri, attrs); err != nil {
		return err
	}
	return idx.inIdx.AddVector(vec, uri, attrs)
}

func (idx *PersistentIndex) Nearest(img image.Image) (string, interface{}, float64, error) {
	return idx.inIdx.Nearest(img)
}

func (idx *PersistentIndex) Remove(f func(embedders.Vector, string, interface{}) bool) ([]string, error) {
	idx.lock.Lock()
	defer idx.lock.Unlock()
	removed, err := idx.inIdx.Remove(f)
	if err != nil || removed == nil {
		return removed, err
	}
	result := idx.db.Where("uri in ?", removed).Delete(&ImgEmbed{})
	if result.Error != nil {
		return nil, fmt.Errorf("failed to remove %d images from db: %w", len(removed), result.Error)
	}
	return removed, nil
}

func (idx *PersistentIndex) GetCount() int { return idx.inIdx.GetCount() }

func NewPersistentIndex(dialector gorm.Dialector, idx Index) (*PersistentIndex, error) {
	db, err := gorm.Open(dialector, &gorm.Config{
		//	Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect db: %w", err)
	}
	err = db.AutoMigrate(&ImgEmbed{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate db: %w", err)
	}

	embeds := make(embeds, 0)
	result := db.Find(&embeds)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to load data from db: %w", result.Error)
	}
	for _, embd := range embeds {
		err := idx.AddVector(embedders.Vector(embd.Vector), embd.URI, embd.Attributes)
		if err != nil {
			return nil, fmt.Errorf("failed to load vectors to index: %w", err)
		}
	}
	return &PersistentIndex{db: db, inIdx: idx}, nil
}
