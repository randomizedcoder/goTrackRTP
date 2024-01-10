package goTrackRTP

// goTracker is for tracking RTP sequence number arrivals

// The general strategy is to store the sequence numbers is a B-tree
// The number of items (.Len()) is the number recieved in the window

// https://github.com/randomizedcoder/goTracker/

// https://pkg.go.dev/container/ring
// https://cs.opensource.google/go/go/+/refs/tags/go1.21.5:src/container/ring/ring.go
// https://cs.opensource.google/go/go/+/refs/tags/go1.21.5:src/container/ring/example_test.go

import (
	"errors"
	"log"

	"github.com/google/btree"
)

const (
	BtreeDegreeCst = 3

	ClearFreeListCst = true

	maxUint16 = ^uint16(0)
)

var (
	ErrUnimplmented = errors.New("ErrUnimplmented")
	ErrPosition     = errors.New("ErrPosition")
)

type Tracker struct {
	b *btree.BTreeG[uint16]

	aw uint16 // aheadWindow
	bw uint16 // behindWindow
	ab uint16 // aheadBuffer
	bb uint16 // behindBuffer

	awPlusAb uint16 // aw + ab
	bwPlusBb uint16 // bw + bb
	Window   uint16 // aw + bw

	debugLevel int
}

type Taxonomy struct {
	Position    int
	Categroy    int
	SubCategory int
	Len         int
	Jump        uint16
}

// Position
const (
	PositionUnknown int = iota
	PositionInit
	PositionAhead
	PositionBehind
	PositionDuplicate
)

// Category
const (
	CategoryUnknown int = iota
	CategoryRestart
	CategoryBuffer
	CategoryWindow
)

// SubCategory
const (
	SubCategoryUnknown int = iota
	SubCategoryNext
	SubCategoryDuplicate
	SubCategoryAlready
	SubCategoryJump
)

type TrackIntToStringMap struct {
	posMap    map[int]string
	catMap    map[int]string
	subCatMap map[int]string
}

func NewMaps() *TrackIntToStringMap {

	pm := make(map[int]string)
	pm[PositionUnknown] = "Unknown"
	pm[PositionInit] = "Init"
	pm[PositionBehind] = "Behind"
	pm[PositionDuplicate] = "Duplicate"

	cm := make(map[int]string)
	cm[CategoryUnknown] = "Unknown"
	cm[CategoryRestart] = "Restart"
	cm[CategoryBuffer] = "Buffer"
	cm[CategoryWindow] = "Window"

	sm := make(map[int]string)
	sm[SubCategoryUnknown] = "Unknown"
	sm[SubCategoryNext] = "Next"
	sm[SubCategoryDuplicate] = "Duplicate"
	sm[SubCategoryAlready] = "Already"
	sm[SubCategoryJump] = "Jump"

	return &TrackIntToStringMap{
		posMap:    pm,
		catMap:    cm,
		subCatMap: sm,
	}
}

// New creates a Tracker
// aw = ahead window
// bw = behind window
// ab = ahead buffer
// bb = behind buffer
func New(aw uint16, bw uint16, ab uint16, bb uint16, debugLevel int) (*Tracker, error) {

	return NewDegree(aw, bw, ab, bb, BtreeDegreeCst, debugLevel)

}

// New creates a Tracker allowing the BTree "degree" to be specified
// See also "degree" or branching factor: https://en.wikipedia.org/wiki/Branching_factor
func NewDegree(aw uint16, bw uint16, ab uint16, bb uint16, degree int, debugLevel int) (*Tracker, error) {

	err := validateNew(aw, bw, ab, bb, degree)
	if err != nil {
		return nil, err
	}

	return &Tracker{
		b: btree.NewG[uint16](degree, isLess),
		//b:        btree.NewOrderedG[uint16](degree),
		aw:       aw,
		bw:       bw,
		ab:       ab,
		bb:       bb,
		awPlusAb: aw + ab,
		bwPlusBb: bw + bb,
		Window:   aw + bw,

		debugLevel: debugLevel,
	}, nil
}

// PacketArrival is the primary packet handling entry point
func (t *Tracker) PacketArrival(seq uint16) (*Taxonomy, error) {

	if t.debugLevel > 10 {
		log.Printf("PacketArrival, seq:%d", seq)
	}

	m, ok := t.b.Max()
	if !ok {
		return t.init(seq)
	}

	if t.debugLevel > 10 {
		log.Printf("PacketArrival, seq:%d, m:%d", seq, m)
	}

	if seq == m {
		return t.positionDuplicate(seq, m)
	}

	if isLessBranchless(seq, m) {
		return t.positionBehind(seq, m)
	} else {
		return t.positionAhead(seq, m)
	}
}

// init is initilizing the data structure on the first packet received
func (t *Tracker) init(seq uint16) (*Taxonomy, error) {

	if t.debugLevel > 10 {
		m, _ := t.b.Max()
		log.Printf("init, seq:%d, t.b.Max():%d, t.b.Len():%d", seq, m, t.b.Len())
	}

	tax := &Taxonomy{}
	tax.Position = PositionInit

	// https://pkg.go.dev/github.com/google/btree#BTree.ReplaceOrInsert
	_, already := t.b.ReplaceOrInsert(seq)
	if already {
		tax.SubCategory = SubCategoryAlready
	}

	if t.debugLevel > 10 {
		m, _ := t.b.Max()
		log.Printf("init, after insert seq:%d, t.b.Max():%d, t.b.Len():%d", seq, m, t.b.Len())
	}

	tax.Len = t.b.Len()

	return tax, nil
}

// positionDuplicate is seq == m
func (t *Tracker) positionDuplicate(seq, m uint16) (*Taxonomy, error) {

	if t.debugLevel > 10 {
		log.Printf("positionDuplicate, seq:%d, t.b.Max():%d, t.b.Len():%d", seq, m, t.b.Len())
	}

	tax := &Taxonomy{}
	tax.Position = PositionDuplicate

	tax.Len = t.b.Len()

	return tax, nil
}

// positionAhead handles seq > Max()2
func (t *Tracker) positionAhead(seq, m uint16) (*Taxonomy, error) {

	if t.debugLevel > 10 {
		log.Printf("positionAhead, seq:%d, t.b.Max():%d, t.b.Len():%d", seq, m, t.b.Len())
	}

	tax := &Taxonomy{}
	tax.Position = PositionAhead

	diff := uint16Diff(seq, m)

	// m < aheadWindow [aw] < categoryBuffer (no op) [aheadBuffer ] < categoryRestart
	if diff > t.awPlusAb {
		return t.categoryRestart(seq, tax)
	} else if diff > t.aw {
		return t.categoryBuffer(seq, tax)
	}
	return t.aheadWindow(seq, m, diff, tax)
}

// positionBehind handles seq < Max()
func (t *Tracker) positionBehind(seq, m uint16) (*Taxonomy, error) {

	if t.debugLevel > 10 {
		log.Printf("positionBehind, seq:%d, t.b.Max():%d, t.b.Len():%d", seq, m, t.b.Len())
	}

	tax := &Taxonomy{}
	tax.Position = PositionBehind

	diff := uint16Diff(seq, m)

	// m < behindWindow [bw] < categoryBuffer (no op) [behindBuffer ] < categoryRestart
	if diff > t.bwPlusBb {

		if t.debugLevel > 10 {
			log.Printf("positionBehind, seq:%d, diff:%d > t.bwPlusBb:%d)", seq, diff, t.bwPlusBb)
		}

		return t.categoryRestart(seq, tax)

	} else if diff > t.bw {

		if t.debugLevel > 10 {
			log.Printf("positionBehind, seq:%d, diff:%d > t.bw:%d", seq, diff, t.bw)
		}

		return t.categoryBuffer(seq, tax)
	}

	if t.debugLevel > 10 {
		log.Println("positionBehind, in window")
	}

	return t.behindWindow(seq, m, diff, tax)
}

// categoryRestart clears the btree and inserts the new seq
// See also: https://pkg.go.dev/github.com/google/btree#BTreeG.Clear
func (t *Tracker) categoryRestart(seq uint16, tax *Taxonomy) (*Taxonomy, error) {

	if t.debugLevel > 10 {
		m, _ := t.b.Max()
		log.Printf("categoryRestart, seq:%d, t.b.Max():%d, t.b.Len():%d", seq, m, t.b.Len())
	}

	tax.Categroy = CategoryRestart

	t.b.Clear(ClearFreeListCst)

	_, already := t.b.ReplaceOrInsert(seq)
	if already {
		tax.SubCategory = SubCategoryAlready
	}

	tax.Len = t.b.Len()

	return tax, nil
}

// categoryBuffer is essentially a no-op
// This is here to make sure we don't reset the window because of some random crazy late/early packet
// With well configured windows this shouldn't happen very often, and if it does maybe your network
// has different latency characteristics than you think?
func (t *Tracker) categoryBuffer(seq uint16, tax *Taxonomy) (*Taxonomy, error) {

	if t.debugLevel > 10 {
		m, _ := t.b.Max()
		log.Printf("categoryBuffer, seq:%d, t.b.Max():%d, t.b.Len():%d", seq, m, t.b.Len())
	}

	tax.Categroy = CategoryBuffer

	tax.Len = t.b.Len()

	return tax, nil
}

// aheadWindow is (hopefully) the most common case
// we need to move the acceptable window forward by clearing items that fall off the back
// See also: https://pkg.go.dev/github.com/google/btree#BTreeG.DescendLessOrEqual
func (t *Tracker) aheadWindow(seq, m uint16, diff uint16, tax *Taxonomy) (*Taxonomy, error) {

	if t.debugLevel > 10 {
		log.Printf("aheadWindow, seq:%d, t.b.Max():%d, t.b.Len():%d", seq, m, t.b.Len())
	}

	tax.Categroy = CategoryWindow
	tax.Jump = diff

	_, duplicate := t.b.ReplaceOrInsert(seq)
	m, _ = t.b.Max()
	if duplicate {
		tax.SubCategory = SubCategoryDuplicate
		if t.debugLevel > 10 {
			log.Printf("aheadWindow, DUPLICATE, seq:%d, t.b.Max():%d, t.b.Len():%d", seq, m, t.b.Len())
		}
	} else if diff == 1 {
		tax.SubCategory = SubCategoryNext
		if t.debugLevel > 10 {
			log.Printf("aheadWindow, diff==1. This is the best outcome! woot woot!")
		}
	} else {
		tax.SubCategory = SubCategoryJump
		if t.debugLevel > 10 {
			log.Printf("aheadWindow, jump:%d", diff)
		}
	}

	if t.debugLevel > 10 {
		m, _ := t.b.Max()
		min, _ := t.b.Min()
		log.Printf("aheadWindow inserted, seq:%d, t.b.Max():%d, t.b.Min():%d, t.b.Len():%d, diff:%d", seq, m, min, t.b.Len(), diff)
	}

	t.deleteItemsFallingOffTheBack(seq)

	tax.Len = t.b.Len()

	return tax, nil
}

// deleteItemsFallingOffTheBack is called by aheadWindow, and deletes
// items falling off the back of the behindWindow
func (t *Tracker) deleteItemsFallingOffTheBack(seq uint16) {

	min, ok := t.b.Min()
	if !ok {
		log.Panicf("aheadWindow Min() not ok:%v, min:%d", ok, min)
	}

	backOfWindow := seq - t.aw - t.bw + 1
	if isLess(min, backOfWindow) {

		// Iterate to clear the items which are falling off the back
		// ( An alternative strategy would be to loop doing deleteMin,
		// but that would be more calls to the btree. )

		var deleted []uint16
		//t.b.DescendLessOrEqual(backOfWindow, func(item uint16) bool {
		t.b.Ascend(func(item uint16) bool {
			if t.debugLevel > 10 {
				log.Printf("aheadWindow, Ascend backOfWindow:%d, min:%d, item:%d", backOfWindow, min, item)
			}
			if isLess(item, backOfWindow) {
				_, ok := t.b.Delete(item)
				if !ok {
					log.Panicf("aheadWindow DescendLessOrEqual Delete not ok:%v", item)
				}
				deleted = append(deleted, item)
				if t.debugLevel > 10 {
					log.Printf("aheadWindow, Ascend backOfWindow:%d, min:%d, deleted item:%d", backOfWindow, min, item)
				}
				return true
			}
			if t.debugLevel > 10 {
				if !isLess(item, backOfWindow) {
					log.Printf("aheadWindow, !isLess(item:%d, backOfWindow:%d)", item, backOfWindow)
				}
			}
			return false
		})

		if t.debugLevel > 10 {
			m, _ := t.b.Max()
			log.Printf("aheadWindow deleted, seq:%d, t.b.Max():%d, t.b.Len():%d, len(deleted):%d, deleted:%v",
				seq, m, t.b.Len(), len(deleted), deleted)
		}
	}
}

// behindWindow handles when the sequence number is within our current
// lookback window
func (t *Tracker) behindWindow(seq, m uint16, diff uint16, tax *Taxonomy) (*Taxonomy, error) {

	if t.debugLevel > 10 {
		log.Printf("behindWindow, seq:%d, t.b.Max():%d, t.b.Len():%d", seq, m, t.b.Len())
	}

	tax.Categroy = CategoryWindow

	_, duplicate := t.b.ReplaceOrInsert(seq)
	m, _ = t.b.Max()
	if duplicate {

		tax.SubCategory = SubCategoryDuplicate

		if t.debugLevel > 10 {
			log.Printf("behindWindow, DUPLICATE, seq:%d, t.b.Max():%d, t.b.Len():%d", seq, m, t.b.Len())
		}

	}
	// We don't track "jump" behind
	// else {
	// 	tax.SubCategory = SubCategoryJump
	// 	if t.debugLevel > 10 {
	// 		log.Printf("behindWindow, jump:%d", diff)
	// 	}
	// }

	tax.Jump = diff

	if t.debugLevel > 10 {
		m, _ := t.b.Max()
		log.Printf("behindWindow inserted, seq:%d, t.b.Max():%d, t.b.Len():%d", seq, m, t.b.Len())
	}

	tax.Len = t.b.Len()

	return tax, nil
}

// Len() returns the current number of items in the btree
// Try not to use this function frequently
func (t *Tracker) Len() int {

	return t.b.Len()

}

// Max() returns the current max item in the btree
// Try not to use this function frequently
func (t *Tracker) Max() uint16 {

	m, _ := t.b.Max()
	return m

}

// Min() returns the current min item in the btree
// Try not to use this function frequently
func (t *Tracker) Min() uint16 {

	m, _ := t.b.Min()
	return m

}

// itemsDescending() iterates descending, returning the list of items
// Try not to use this function frequently ( expensive )
func (t *Tracker) itemsDescending() (items []uint16) {
	t.b.Descend(func(item uint16) bool {
		items = append(items, item)
		return true
	})
	return items
}
