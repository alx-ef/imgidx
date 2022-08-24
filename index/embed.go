package index

import (
	"gonum.org/v1/gonum/spatial/kdtree"
	"gorm.io/gorm"
)

type ImgEmbed struct {
	gorm.Model
	URI        string       `gorm:"unique"`
	Vector     kdtree.Point `gorm:"serializer:json"`
	Attributes interface{}  `gorm:"serializer:json"`
	_          struct{}     `gorm:"-"`
}

func (embd ImgEmbed) Compare(c kdtree.Comparable, d kdtree.Dim) float64 {
	return embd.Vector.Compare(c.(ImgEmbed).Vector, d)
}

// Dims returns the number of dimensions described by the receiver.
func (embd ImgEmbed) Dims() int { return len(embd.Vector) }

// Distance returns the squared Euclidean distance between c and the receiver. The
// concrete type of c must be Point.
func (embd ImgEmbed) Distance(c kdtree.Comparable) float64 {
	return embd.Vector.Distance(c.(ImgEmbed).Vector)
}

// images is a collection of the image type that satisfies kdtree.Interface.
type embeds []ImgEmbed

func (e embeds) Index(i int) kdtree.Comparable         { return e[i] }
func (e embeds) Len() int                              { return len(e) }
func (e embeds) Pivot(d kdtree.Dim) int                { return plane{d, e}.Pivot() }
func (e embeds) Slice(start, end int) kdtree.Interface { return e[start:end] }

// plane is required to help places.
type plane struct {
	kdtree.Dim
	embeds
}

func (p plane) Less(i, j int) bool                     { return p.embeds[i].Vector[p.Dim] < p.embeds[j].Vector[p.Dim] }
func (p plane) Pivot() int                             { return kdtree.Partition(p, kdtree.MedianOfMedians(p)) }
func (p plane) Slice(start, end int) kdtree.SortSlicer { p.embeds = p.embeds[start:end]; return p }
func (p plane) Swap(i, j int)                          { p.embeds[i], p.embeds[j] = p.embeds[j], p.embeds[i] }
