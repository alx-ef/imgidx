package imgidx

import "gonum.org/v1/gonum/spatial/kdtree"

type embed struct {
	kdtree.Point
	uri   string
	attrs interface{}
}

func (p embed) Compare(c kdtree.Comparable, d kdtree.Dim) float64 {
	return p.Point.Compare(c.(embed).Point, d)
}

// Dims returns the number of dimensions described by the receiver.
func (p embed) Dims() int { return len(p.Point) }

// Distance returns the squared Euclidean distance between c and the receiver. The
// concrete type of c must be Point.
func (p embed) Distance(c kdtree.Comparable) float64 { return p.Point.Distance(c.(embed).Point) }

// images is a collection of the image type that satisfies kdtree.Interface.
type embeds []embed

func (p embeds) Index(i int) kdtree.Comparable         { return p[i] }
func (p embeds) Len() int                              { return len(p) }
func (p embeds) Pivot(d kdtree.Dim) int                { return plane{d, p}.Pivot() }
func (p embeds) Slice(start, end int) kdtree.Interface { return p[start:end] }

// plane is required to help places.
type plane struct {
	kdtree.Dim
	embeds
}

func (p plane) Less(i, j int) bool                     { return p.embeds[i].Point[p.Dim] < p.embeds[j].Point[p.Dim] }
func (p plane) Pivot() int                             { return kdtree.Partition(p, kdtree.MedianOfMedians(p)) }
func (p plane) Slice(start, end int) kdtree.SortSlicer { p.embeds = p.embeds[start:end]; return p }
func (p plane) Swap(i, j int)                          { p.embeds[i], p.embeds[j] = p.embeds[j], p.embeds[i] }
