package wiki

import (
	"sort"
	"strings"
	"time"
)

// Sortable is the interface that allows quiki to sort wiki resources.
type Sortable interface {
	SortInfo() SortInfo
}

// SortInfo is the data returned from Sortable items for sorting wiki resources.
type SortInfo struct {
	Title      string
	Author     string
	Created    time.Time
	Modified   time.Time
	Dimensions []int
}

// SortFunc is a type for functions that can sort items.
type SortFunc func(p, q Sortable) bool

// SortTitle is a SortFunc for sorting items alphabetically by title.
func SortTitle(p, q Sortable) bool {
	return strings.ToLower(p.SortInfo().Title) < strings.ToLower(q.SortInfo().Title)
}

// SortAuthor is a SortFunc for sorting items alphabetically by author.
func SortAuthor(p, q Sortable) bool {
	return strings.ToLower(p.SortInfo().Author) < strings.ToLower(q.SortInfo().Author)
}

// SortCreated is a SortFunc for sorting items by creation time.
func SortCreated(p, q Sortable) bool {
	return p.SortInfo().Created.Before(q.SortInfo().Created)
}

// SortModified is a SortFunc for sorting items by modification time.
func SortModified(p, q Sortable) bool {
	return p.SortInfo().Modified.Before(q.SortInfo().Modified)
}

// SortDimensions is a SortFunc for sorting images by their dimensions.
func SortDimensions(p, q Sortable) bool {
	d1, d2 := p.SortInfo().Dimensions, q.SortInfo().Dimensions
	if d1 == nil || d2 == nil {
		return false
	}
	product1, product2 := d1[0]*d1[1], d2[0]*d2[1]
	return product1 < product2
}

// itemSorter implements the Sort interface, sorting the changes within.
type itemSorter struct {
	items []Sortable
	less  []SortFunc
}

// Sort sorts the argument slice according to the less functions passed to itemsOrderedBy.
func (ps *itemSorter) Sort(items []Sortable) {
	ps.items = items
	sort.Sort(ps)
}

// sorter returns a Sorter that sorts using the less functions, in order.
// Call its Sort method to sort the data.
func sorter(items []Sortable, less ...SortFunc) *itemSorter {
	return &itemSorter{items, less}
}

// Len is part of sort.Interface.
func (ps *itemSorter) Len() int {
	return len(ps.items)
}

// Swap is part of sort.Interface.
func (ps *itemSorter) Swap(i, j int) {
	ps.items[i], ps.items[j] = ps.items[j], ps.items[i]
}

// Less is part of sort.Interface. It is implemented by looping along the
// less functions until it finds a comparison that discriminates between
// the two items (one is less than the other). Note that it can call the
// less functions twice per call. We could change the functions to return
// -1, 0, 1 and reduce the number of calls for greater efficiency: an
// exercise for the reader.
func (ps *itemSorter) Less(i, j int) bool {
	p, q := ps.items[i], ps.items[j]
	// Try all but the last comparison.
	var k int
	for k = 0; k < len(ps.less)-1; k++ {
		less := ps.less[k]
		switch {
		case less(p, q):
			// p < q, so we have a decision.
			return true
		case less(q, p):
			// p > q, so we have a decision.
			return false
		}
		// p == q; try the next comparison.
	}
	// All comparisons to here said "equal", so just return whatever
	// the final comparison reports.
	return ps.less[k](p, q)
}
