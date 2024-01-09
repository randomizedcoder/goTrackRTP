package goTrackRTP

import (
	"reflect"
	"testing"
	_ "unsafe"
)

// https://github.com/randomizedcoder/goTrackRTP/

// See also: https://dave.cheney.net/2019/05/07/prefer-table-driven-tests

func BenchmarkIsLess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			isLess(uint16(i), uint16(i+5))
		} else {
			isLess(uint16(i), uint16(i-5))
		}
	}
}

func BenchmarkIsGreaterBranchless(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			isLessBranchless(uint16(i), uint16(i+5))
		} else {
			isLessBranchless(uint16(i), uint16(i-5))
		}
	}
}

func BenchmarkIsGreaterBranch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			isLessBranch(uint16(i), uint16(i+5))
		} else {
			isLessBranch(uint16(i), uint16(i-5))
		}
	}
}

func TestIsLess(t *testing.T) {
	type test struct {
		m    uint16
		seq  uint16
		want bool
	}

	tests := []test{
		// obvious
		{0, 1, false},
		{1, 0, true},
		{100, 101, false},
		{101, 100, true},
		{65535, 65534, true},
		// less obvious
		{0, 65535, true},
		{65535, 0, false},
		{1, 65535, true},
		{65535, 1, false},
		{10, 65535, true},
		{65535, 10, false},
		{100, 65535, true},
		{65535, 100, false},
		{0, maxUint16 / 2, false},
		{maxUint16 / 2, 0, true},
		// more for good measure
		{0, 65000, true},
		{65000, 0, false},
		{1, 65000, true},
		{65000, 1, false},
		{10, 65000, true},
		{65000, 10, false},
		{100, 65000, true},
		{65000, 100, false},
		// same
		{0, 0, false},
		{1000, 1000, false},
		{maxUint16 / 2, maxUint16 / 2, false},
		{65535, 65535, false},
	}

	for i, tc := range tests {

		t.Logf("%s i:%d, m:%v, seq:%v\n", t.Name(), i, tc.m, tc.seq)

		got := isLess(tc.seq, tc.m)

		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("test: %d, expected: %v, got: %v", i, tc.want, got)
		}
	}
}

func TestIsLessBranchless(t *testing.T) {
	type test struct {
		m    uint16
		seq  uint16
		want bool
	}

	tests := []test{
		// obvious
		{0, 1, false},
		{1, 0, true},
		{100, 101, false},
		{101, 100, true},
		{65535, 65534, true},
		// less obvious
		{0, 65535, true},
		{65535, 0, false},
		{1, 65535, true},
		{65535, 1, false},
		{10, 65535, true},
		{65535, 10, false},
		{100, 65535, true},
		{65535, 100, false},
		{0, maxUint16 / 2, false},
		{maxUint16 / 2, 0, true},
		// more for good measure
		{0, 65000, true},
		{65000, 0, false},
		{1, 65000, true},
		{65000, 1, false},
		{10, 65000, true},
		{65000, 10, false},
		{100, 65000, true},
		{65000, 100, false},
		// same
		{0, 0, false},
		{1000, 1000, false},
		{maxUint16 / 2, maxUint16 / 2, false},
		{65535, 65535, false},
	}

	for i, tc := range tests {

		t.Logf("%s i:%d, m:%v, seq:%v\n", t.Name(), i, tc.m, tc.seq)

		got := isLessBranchless(tc.seq, tc.m)

		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("test: %d, expected: %v, got: %v", i, tc.want, got)
		}
	}
}

func TestIsLessBranch(t *testing.T) {
	type test struct {
		m    uint16
		seq  uint16
		want bool
	}

	tests := []test{
		// obvious
		{0, 1, false},
		{1, 0, true},
		{100, 101, false},
		{101, 100, true},
		{65535, 65534, true},
		// less obvious
		{0, 65535, true},
		{65535, 0, false},
		{1, 65535, true},
		{65535, 1, false},
		{10, 65535, true},
		{65535, 10, false},
		{100, 65535, true},
		{65535, 100, false},
		{0, maxUint16 / 2, false},
		{maxUint16 / 2, 0, true},
		// more for good measure
		{0, 65000, true},
		{65000, 0, false},
		{1, 65000, true},
		{65000, 1, false},
		{10, 65000, true},
		{65000, 10, false},
		{100, 65000, true},
		{65000, 100, false},
		// same
		{0, 0, false},
		{1000, 1000, false},
		{maxUint16 / 2, maxUint16 / 2, false},
		{65535, 65535, false},
	}

	for i, tc := range tests {

		t.Logf("%s i:%d, m:%v, seq:%v\n", t.Name(), i, tc.m, tc.seq)

		got := isLessBranch(tc.seq, tc.m)

		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("test: %d, expected: %v, got: %v", i, tc.want, got)
		}
	}
}

func TestDiff(t *testing.T) {
	type test struct {
		m    uint16
		seq  uint16
		want uint16
	}

	tests := []test{
		{0, 1, 1},
		{1, 0, 1},
		// wrapping
		{0, 65535, 1},
		{65535, 0, 1},
		{1, 65535, 2},
		{65535, 1, 2},
		{10, 65535, 11},
		{65535, 10, 11},
		// obvious
		{0, 1, 1},
		{1, 0, 1},
		{100, 101, 1},
		{101, 100, 1},
		{65534, 65535, 1},
		{65535, 65534, 1},
		{1, 65535, 2},
		{10, 65535, 11},
		{100, 65535, 101},
		{0, maxUint16 / 2, maxUint16 / 2},
		{maxUint16 / 2, 0, maxUint16 / 2},
		//{0, 65000, 65000}, // fails
		{1, 65000, 537},
		{10, 65000, 546},
		{100, 65000, 636},
		// // same
		{0, 0, 0},
		{1000, 1000, 0},
		{maxUint16 / 2, maxUint16 / 2, 0},
		{65535, 65535, 0},
	}

	// I was trying to work out a branchless version, but didn't get there yet
	// can't work out how to iterate function :(
	// var fns []func(s1, s2 uint16)
	// fns = append(fns, absoluteDifference)
	// fns = append(fns, uint16Diff)
	//fns := []string{"absoluteDifference", "uint16Diff"}
	//fns := []string{"uint16Diff"}

	for i, tc := range tests {

		//for f, fn := range fns {

		t.Logf("%s i:%d, m:%v, seq:%v, want:%d\n", t.Name(), i, tc.m, tc.seq, tc.want)
		//t.Logf("%s i:%d, fn:%s, m:%v, seq:%v, want:%d\n", t.Name(), i, fn, tc.m, tc.seq, tc.want)

		//got := fn(tc.m, tc.seq)

		// var got uint16
		// if f == 1 {
		// 	got = absoluteDifference(tc.m, tc.seq)
		// } else {
		// 	got = uint16Diff(tc.m, tc.seq)
		// }
		got := uint16Diff(tc.m, tc.seq)

		t.Logf("%s i:%d, m:%v, seq:%v, want:%d, got:%d, deepequal:%t\n", t.Name(), i, tc.m, tc.seq, tc.want, got, reflect.DeepEqual(tc.want, got))

		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("test: %d, expected: %v, got: %v", i, tc.want, got)
		}
		//}
	}
}
