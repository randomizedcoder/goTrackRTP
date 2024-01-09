package goTrackRTP

// https://github.com/randomizedcoder/goTracker/

// Math related functions ( sequence wrap handling )

// Quote:
// This function works by comparing the two numbers and their difference.
// If their difference is less than 1/2 the maximum sequence number value,
// then they must be close together - so we just check if one is greater
// than the other, as usual.
// However, if they are far apart, their difference will be greater than 1/2
// the max sequence, then we paradoxically consider the sequence number more
// recent if it is less than the current sequence number.
// https://gafferongames.com/post/reliability_ordering_and_congestion_avoidance_over_udp/

// See also: https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go

// Using Less because the btree library uses less
// https://github.com/google/btree/blob/v1.1.2/btree_generic.go#L135

// isLess(seq, m uint16) is !isGreater
// NOTE seq is first argument!!  We swap them here
func isLess[T uint16](s1, s2 uint16) bool {
	//log.Printf("isLess, s1:%d, s2:%d", s1, s2)
	b := isLessBranchless(s1, s2)
	//log.Printf("isLess, isLessBranchless(seq, m):%t", b)
	return b
}

// isLessBranchless is a non-banching (if-ess) version to find less that handles sequence wrapping
func isLessBranchless(s1, s2 uint16) bool {

	diff := int(s1) - int(s2)

	diff += int(maxUint16/2) + 1
	diff &= int(maxUint16)

	return diff > 0 && diff <= int(maxUint16/2)
}

// isLessBranch is a banching (if) version to find less that handles sequence wrapping
func isLessBranch(s1, s2 uint16) bool {

	if s1 < s2 {
		return s2-s1 <= maxUint16/2
	} else {
		return s1-s2 > maxUint16/2
	}
}

// uint16Diff returns difference in uint16 sequence handling wrapping
func uint16Diff(s1, s2 uint16) uint16 {

	//log.Printf("uint16Diff, m:%d, seq:%d", m, seq)

	if s1 == s2 {
		return 0
	}

	var abs uint16
	if s1 < s2 {
		abs = s2 - s1
	} else {
		abs = s1 - s2
	}

	if abs > maxUint16/2 {
		return maxUint16 - abs + 1
	}

	//log.Printf("uint16Diff, s1:%d, s2:%d, abs:%d", s1, s2, abs)

	return abs
}

// // isGreater determines if seq is ahead of m
// func isGreater[T uint16](m, seq uint16) bool {
// 	return isGreaterBranchless(m, seq)
// }
// // isGreaterBranchless determines if seq is ahead of m
// // branchless version
// func isGreaterBranchless[T uint16](m, seq uint16) bool {

// 	diff := int(m) - int(seq)
// 	diff += int(maxUint16/2) + 1
// 	diff &= int(maxUint16)

// 	return diff > 0 && diff <= int(maxUint16/2)
// }

// // isGreater determines if seq is ahead of m
// func isGreaterBranch[T uint16](m, seq uint16) bool {
// 	if m < seq {
// 		return seq-m <= maxUint16/2
// 	} else {
// 		return m-seq > maxUint16/2
// 	}
// }
