package goTrackRTP

// https://github.com/randomizedcoder/goTrackRTP/

// See also: https://dave.cheney.net/2019/05/07/prefer-table-driven-tests

import (
	"math"
	"os"
	"reflect"
	"testing"
	_ "unsafe"
)

// unsafe for fastrand

const (
	debugLevelCst        = 11
	WindowSizeTestingCst = 100
)

// unsafe for the FastRand()
//_ "unsafe"
// //go:linkname FastRand runtime.fastrand
// func FastRand() uint32

// // https://cs.opensource.google/go/go/+/master:src/runtime/stubs.go;l=151?q=FastRandN&ss=go%2Fgo
// // https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/

// //go:linkname FastRandN runtime.fastrandn
// func FastRandN(n uint32) uint32

func TestTrackerInit(t *testing.T) {

	type test struct {
		aw     uint16
		bw     uint16
		ab     uint16
		bb     uint16
		err    error
		Window uint16
		Len    int
	}

	tests := []test{
		// basic
		{10, 10, 10, 10, nil, 10 + 10, 0},
		{100, 100, 100, 100, nil, 200, 0},
		{1500, 1500, 1500, 1500, nil, 1500 + 1500, 0},
		{10, 20, 10, 10, nil, 10 + 20, 0},

		// errors
		{1, 10, 10, 10, ErrWindowAWMin, 0, 0},
		{1501, 10, 10, 10, ErrWindowAWMax, 0, 0},
		{10, 1, 10, 10, ErrWindowBWMin, 0, 0},
		{10, 1501, 10, 10, ErrWindowBWMax, 0, 0},
		{10, 10, 1, 10, ErrWindowABMin, 0, 0},
		{10, 10, 1501, 10, ErrWindowABMax, 0, 0},
		{10, 10, 10, 1, ErrWindowBBMin, 0, 0},
		{10, 10, 10, 1501, ErrWindowBBMax, 0, 0},

		// window too small
		{0, 0, 0, 0, ErrWindowAWMin, 0, 0},
		{1, 1, 1, 1, ErrWindowAWMin, 0, 0},
		{3, 3, 3, 3, ErrWindowAWMin, 0, 0},
		{1501, 1501, 1501, 1501, ErrWindowAWMax, 0, 0},
		{1501, 1501, 1501, 1501, ErrWindowAWMax, 0, 0},
	}

	for i, tc := range tests {

		t.Logf("%s i:%d, tc: %v\n", t.Name(), i, tc)

		tr, err := New(tc.aw, tc.bw, tc.ab, tc.bb, debugLevelCst)

		if err != tc.err {
			t.Fatalf("%s, err:%v != tc.err:%v", t.Name(), err, tc.err)
		}

		if err == nil {
			if tr.Window != tc.Window {
				t.Fatalf("%s, tr.Window:%d != tc.Window:%d", t.Name(), tr.Window, tc.Window)
			}

			if tr.Len() != tc.Len {
				t.Fatalf("%s, etr.Len():%d != tc.Len:%d", t.Name(), tr.Len(), tc.Len)
			}
		}
	}
}

func TestTrackerWindow(t *testing.T) {

	type test struct {
		aw          uint16
		bw          uint16
		ab          uint16
		bb          uint16
		m           uint16
		seq         uint16
		err         error
		Window      uint16
		Len         int
		Max         uint16
		Jump        uint16
		Position    int
		Category    int
		SubCategory int
	}

	tests := []test{
		// position duplicate
		{10, 10, 10, 10, 0, 0, nil, 10 + 10, 1, 0, 0, PositionDuplicate, CategoryUnknown, SubCategoryUnknown},
		{10, 10, 10, 10, 1, 1, nil, 10 + 10, 1, 1, 0, PositionDuplicate, CategoryUnknown, SubCategoryUnknown},
		{10, 10, 10, 10, maxUint16, maxUint16, nil, 10 + 10, 1, maxUint16, 0, PositionDuplicate, CategoryUnknown, SubCategoryUnknown},
		{10, 10, 10, 10, maxUint16 - 1, maxUint16 - 1, nil, 10 + 10, 1, maxUint16 - 1, 0, PositionDuplicate, CategoryUnknown, SubCategoryUnknown},
		{100, 100, 100, 100, 0, 0, nil, 100 + 100, 1, 0, 0, PositionDuplicate, CategoryUnknown, SubCategoryUnknown},
		// position ahead - window
		{10, 10, 10, 10, 0, 1, nil, 10 + 10, 2, 1, 1, PositionAhead, CategoryWindow, SubCategoryNext}, // failing here!  max not updating
		{10, 10, 10, 10, 0, 10, nil, 10 + 10, 2, 10, 10, PositionAhead, CategoryWindow, SubCategoryJump},
		{100, 100, 100, 100, 0, 1, nil, 100 + 100, 2, 1, 1, PositionAhead, CategoryWindow, SubCategoryNext},
		{100, 100, 100, 100, 0, 10, nil, 100 + 100, 2, 10, 10, PositionAhead, CategoryWindow, SubCategoryJump},
		{100, 100, 100, 100, 0, 100, nil, 100 + 100, 2, 100, 100, PositionAhead, CategoryWindow, SubCategoryJump},
		// position ahead - buffer
		{10, 10, 10, 10, 0, 11, nil, 10 + 10, 1, 0, 0, PositionAhead, CategoryBuffer, SubCategoryUnknown},
		{10, 10, 10, 10, 0, 15, nil, 10 + 10, 1, 0, 0, PositionAhead, CategoryBuffer, SubCategoryUnknown},
		{10, 10, 10, 10, 0, 19, nil, 10 + 10, 1, 0, 0, PositionAhead, CategoryBuffer, SubCategoryUnknown},
		{10, 10, 10, 10, 0, 20, nil, 10 + 10, 1, 0, 0, PositionAhead, CategoryBuffer, SubCategoryUnknown},
		{100, 100, 100, 100, 0, 101, nil, 100 + 100, 1, 0, 0, PositionAhead, CategoryBuffer, SubCategoryUnknown},
		{100, 100, 100, 100, 0, 200, nil, 100 + 100, 1, 0, 0, PositionAhead, CategoryBuffer, SubCategoryUnknown},
		// position ahead - restart
		{10, 10, 10, 10, 0, 21, nil, 10 + 10, 1, 21, 0, PositionAhead, CategoryRestart, SubCategoryUnknown},
		{10, 10, 10, 10, 0, 100, nil, 10 + 10, 1, 100, 0, PositionAhead, CategoryRestart, SubCategoryUnknown},
		{10, 10, 10, 10, 0, 1000, nil, 10 + 10, 1, 1000, 0, PositionAhead, CategoryRestart, SubCategoryUnknown},
		{100, 100, 100, 100, 0, 201, nil, 100 + 100, 1, 201, 0, PositionAhead, CategoryRestart, SubCategoryUnknown},
		{100, 100, 100, 100, 0, 1000, nil, 100 + 100, 1, 1000, 0, PositionAhead, CategoryRestart, SubCategoryUnknown},
		// position ahead - window

		// - note because we only insert x2 in this test, we can't test PositionAhead + SubCategoryDuplicate
		// position behind - window
		{10, 10, 10, 10, 0, maxUint16, nil, 10 + 10, 2, 0, 1, PositionBehind, CategoryWindow, SubCategoryUnknown},
		{10, 10, 10, 10, 0, maxUint16 - 5, nil, 10 + 10, 2, 0, 6, PositionBehind, CategoryWindow, SubCategoryUnknown},
		{10, 10, 10, 10, 0, maxUint16 - 9, nil, 10 + 10, 2, 0, 10, PositionBehind, CategoryWindow, SubCategoryUnknown},
		{100, 100, 100, 100, 0, maxUint16, nil, 100 + 100, 2, 0, 1, PositionBehind, CategoryWindow, SubCategoryUnknown},
		// position behind - buffer
		{10, 10, 10, 10, 0, maxUint16 - 10, nil, 10 + 10, 1, 0, 0, PositionBehind, CategoryBuffer, SubCategoryUnknown},
		{10, 10, 10, 10, 0, maxUint16 - 11, nil, 10 + 10, 1, 0, 0, PositionBehind, CategoryBuffer, SubCategoryUnknown},
		{10, 10, 10, 10, 0, maxUint16 - 19, nil, 10 + 10, 1, 0, 0, PositionBehind, CategoryBuffer, SubCategoryUnknown},
		{100, 100, 100, 100, 0, maxUint16 - 100, nil, 100 + 100, 1, 0, 0, PositionBehind, CategoryBuffer, SubCategoryUnknown},
		{100, 100, 100, 100, 0, maxUint16 - 199, nil, 100 + 100, 1, 0, 0, PositionBehind, CategoryBuffer, SubCategoryUnknown},
		// position behind - restart
		{10, 10, 10, 10, 0, maxUint16 - 20, nil, 10 + 10, 1, maxUint16 - 20, 0, PositionBehind, CategoryRestart, SubCategoryUnknown},
		{10, 10, 10, 10, 0, maxUint16 - 100, nil, 10 + 10, 1, maxUint16 - 100, 0, PositionBehind, CategoryRestart, SubCategoryUnknown},
		{10, 10, 10, 10, 0, maxUint16 - 1000, nil, 10 + 10, 1, maxUint16 - 1000, 0, PositionBehind, CategoryRestart, SubCategoryUnknown},
		{100, 100, 100, 100, 0, maxUint16 - 200, nil, 100 + 100, 1, maxUint16 - 200, 0, PositionBehind, CategoryRestart, SubCategoryUnknown},
	}

	for i, tc := range tests {

		t.Logf("%s i:%d, tc: %v\n", t.Name(), i, tc)

		tr, err := New(tc.aw, tc.bw, tc.ab, tc.bb, debugLevelCst)
		if err != tc.err {
			t.Fatalf("%s, err:%v != tc.err:%v", t.Name(), err, tc.err)
		}

		_, e := tr.PacketArrival(tc.m)
		if e != nil {
			t.Fatalf("%s, err != nil:%v", t.Name(), e)
		}

		tax, et := tr.PacketArrival(tc.seq)
		if et != nil {
			t.Fatalf("%s, err != nil:%v", t.Name(), et)
		}

		if !reflect.DeepEqual(tr.Window, tc.Window) {
			t.Fatalf("%s, test:%d !reflect.DeepEqual(tr.Window:%v, tc.Window:%v)", t.Name(), i, tr.Window, tc.Window)
		}

		if !reflect.DeepEqual(tax.Len, tc.Len) {
			t.Fatalf("%s, test:%d !reflect.DeepEqual(tax.Len:%v, tc.Len:%v)", t.Name(), i, tax.Len, tc.Len)
		}

		if !reflect.DeepEqual(tr.Max(), tc.Max) {
			t.Fatalf("%s, test:%d !reflect.DeepEqual(tr.Max():%v, tc.Max:%v)", t.Name(), i, tr.Max(), tc.Max)
		}

		if !reflect.DeepEqual(tax.Jump, tc.Jump) {
			t.Fatalf("%s, test:%d !reflect.DeepEqual(tax.Jump:%v, tc.Jump:%v)", t.Name(), i, tax.Jump, tc.Jump)
		}

		if !reflect.DeepEqual(tax.Position, tc.Position) {
			t.Fatalf("%s, test:%d !reflect.DeepEqual(tax.Position:%v, tc.Position:%v)", t.Name(), i, tax.Position, tc.Position)
		}
		if !reflect.DeepEqual(tax.Categroy, tc.Category) {
			t.Fatalf("%s, test:%d !reflect.DeepEqual(tax.Categroy:%v, tc.Category:%v)", t.Name(), i, tax.Categroy, tc.Category)
		}
		if !reflect.DeepEqual(tax.SubCategory, tc.SubCategory) {
			t.Fatalf("%s, test:%d !reflect.DeepEqual(tax.SubCategory:%v, tc.SubCategory:%v)", t.Name(), i, tax.SubCategory, tc.SubCategory)
		}
	}
}

func TestLongRunningWindow(t *testing.T) {

	type test struct {
		aw    uint16
		bw    uint16
		ab    uint16
		bb    uint16
		dl    int
		start uint16
		err   error
		loops int64
		Len   int
	}

	tests := []test{
		{10, 10, 10, 10, 11, 0, nil, 21, 20},
		{10, 10, 10, 10, 11, 0, nil, 41, 20},
		{10, 10, 10, 10, 11, maxUint16 - 10, nil, 41, 20},
		{100, 100, 100, 100, 11, 0, nil, 201, 200},
		{100, 100, 100, 100, 11, 0, nil, 401, 200},
		{100, 100, 100, 100, 11, maxUint16 - 100, nil, 401, 200},
		{1000, 1000, 1000, 1000, 11, 0, nil, 2001, 2000},
		{1000, 1000, 1000, 1000, 11, 0, nil, 4001, 2000},
		{1000, 1000, 1000, 1000, 11, maxUint16 - 1000, nil, 4001, 2000},
		// long
		{10, 10, 10, 10, 0, 0, nil, (math.MaxInt32 * 2) + 21, 20},
		{100, 100, 100, 100, 0, 0, nil, (math.MaxInt32 * 2) + 201, 200},
		{100, 100, 100, 100, 11, 0, nil, (math.MaxInt32 * 2) + 201, 200},
	}

	if os.Getenv("LONG") != "true" {
		t.Skip("Skipping long test.  Set 'LONG=true' env var to run this")
	}

	for i, tc := range tests {

		t.Logf("%s i:%d, tc: %v\n", t.Name(), i, tc)

		tr, err := New(tc.aw, tc.bw, tc.ab, tc.bb, tc.dl)
		if err != tc.err {
			t.Fatalf("%s, err:%v != tc.err:%v", t.Name(), err, tc.err)
		}

		var tax *Taxonomy
		var e error
		var loops int64
		var j uint16 = tc.start
		for {
			if tc.dl > 10 {
				t.Logf("%s i:%d, tc: %v, j:%d, loops:%d\n", t.Name(), i, tc, j, loops)
			}
			tax, e = tr.PacketArrival(j)
			if e != nil {
				t.Fatalf("%s, e != nil:%v", t.Name(), e)
			}
			j++
			loops++
			if loops > tc.loops {
				t.Logf("loops:%d > tc.Loops:%d, tr.Max():%d, tax.Len:%d, tr.Min():%d", loops, tc.loops, tr.Max(), tax.Len, tr.Min())
				if tc.dl > 110 {
					t.Logf("items:%v", tr.itemsDescending())
				}
				break
			}
		}
		if !reflect.DeepEqual(tax.Len, tc.Len) {
			t.Fatalf("%s, test:%d !reflect.DeepEqual(tax.Len:%v, tc.Len:%v)", t.Name(), i, tax.Len, tc.Len)
		}
	}
}

func TestLongRunningJumps(t *testing.T) {

	type test struct {
		aw          uint16
		bw          uint16
		ab          uint16
		bb          uint16
		dl          int
		start       uint16
		err         error
		loops       int64
		MaxRandJump uint32
		Len         int
	}

	debugL := 0

	tests := []test{
		{10, 10, 10, 10, debugL, 0, nil, 21, 10, 20},
		{10, 10, 10, 10, debugL, 0, nil, 41, 10, 20},
		{10, 10, 10, 10, debugL, maxUint16 - 10, nil, 41, 10, 20},
		{100, 100, 100, 100, debugL, 0, nil, 201, 100, 200},
		{100, 100, 100, 100, debugL, 0, nil, 401, 100, 200},
		{100, 100, 100, 100, debugL, maxUint16 - 100, nil, 401, 100, 200},
		{1000, 1000, 1000, 1000, debugL, 0, nil, 2001, 1000, 2000},
		{1000, 1000, 1000, 1000, debugL, 0, nil, 4001, 1000, 2000},
		{1000, 1000, 1000, 1000, debugL, maxUint16 - 1000, nil, 4001, 1000, 2000},
		// long
		{10, 10, 10, 10, 0, 0, nil, (int64(maxUint16) * 3) + 21, 10, 20},
		{100, 100, 100, 100, 0, 0, nil, (int64(maxUint16) * 3) + 201, 100, 200},
		{100, 100, 100, 100, 0, 0, nil, (int64(maxUint16) * 3) + 201, 100, 200},
	}

	if os.Getenv("LONG") != "true" {
		t.Skip("Skipping long test.  Set 'LONG=true' env var to run this")
	}

	for i, tc := range tests {

		t.Logf("%s i:%d, tc: %v\n", t.Name(), i, tc)

		tr, err := New(tc.aw, tc.bw, tc.ab, tc.bb, tc.dl)
		if err != tc.err {
			t.Fatalf("%s, err:%v != tc.err:%v", t.Name(), err, tc.err)
		}

		var tax *Taxonomy
		var e error
		var loops int64
		var j uint16 = tc.start
		var dupSent int
		var dup int
		for {
			if tc.dl > 10 {
				t.Logf("%s i:%d, tc: %v, j:%d, loops:%d", t.Name(), i, tc, j, loops)
			}
			tax, e = tr.PacketArrival(j)
			if e != nil {
				t.Fatalf("%s, e != nil:%v", t.Name(), e)
			}
			if loops > int64(tc.MaxRandJump) {
				if tax.SubCategory != SubCategoryNext {
					t.Fatalf("%s, test:%d tax.SubCategory:%v != SubCategoryNext:%v", t.Name(), i, tax.SubCategory, SubCategoryNext)
				}

				// send a duplicate in the behind window
				r := uint16(FastRandN(tc.MaxRandJump-1) + 1) // FastRandN can return zero (0)
				if tc.dl > 10 {
					t.Logf("%s i:%d, r:%d, j - r:%d", t.Name(), i, r, j-r)
				}
				tax, e = tr.PacketArrival(j - r)
				if e != nil {
					t.Fatalf("%s, e != nil:%v", t.Name(), e)
				}
				dupSent++
				if tax.SubCategory == SubCategoryDuplicate {
					dup++
				} else {
					t.Fatalf("%s, test:%d tax.SubCategory != SubCategoryDuplicate", t.Name(), i)
				}
			}

			j++
			loops++
			if loops > tc.loops {
				t.Logf("loops:%d > tc.Loops:%d, tr.Max():%d, tax.Len:%d, tr.Min():%d", loops, tc.loops, tr.Max(), tax.Len, tr.Min())
				if tc.dl > 110 {
					t.Logf("items:%v", tr.itemsDescending())
				}
				break
			}

			if tc.dl > 10 {
				if loops%int64(maxUint16) == 0 {
					t.Logf("loops:%d > tc.Loops:%d, tr.Max():%d, tax.Len:%d, tr.Min():%d", loops, tc.loops, tr.Max(), tax.Len, tr.Min())
				}
			}
		}
		// if !reflect.DeepEqual(tax.Len, tc.Len) {
		// 	t.Fatalf("%s, test:%d !reflect.DeepEqual(tax.Len:%v, tc.Len:%v)", t.Name(), i, tax.Len, tc.Len)
		// }
		if dupSent != dup {
			t.Fatalf("%s, test:%d dupSent:%d != dup:%d", t.Name(), i, dupSent, dup)
		} else {
			t.Logf("%s i:%d, tc: %v, j:%d, loops:%d, duplicate test succeeded! dup:%d", t.Name(), i, tc, j, loops, dup)
		}
	}
}

//go:linkname FastRand runtime.fastrand
func FastRand() uint32

// https://cs.opensource.google/go/go/+/master:src/runtime/stubs.go;l=151?q=FastRandN&ss=go%2Fgo
// https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/

//go:linkname FastRandN runtime.fastrandn
func FastRandN(n uint32) uint32
