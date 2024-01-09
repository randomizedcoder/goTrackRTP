package goTrackRTP

// Input validation

import (
	"errors"
	"log"
)

const (
	MinWindowCst = 3
	MaxWindowCst = 1500
	MinDegree    = 2
	MaxDegree    = 10
)

var (
	ErrWindowAWMin     = errors.New("ErrWindow window ahead min")
	ErrWindowAWMax     = errors.New("ErrWindow window ahead max")
	ErrWindowBWMin     = errors.New("ErrWindow window behind min")
	ErrWindowBWMax     = errors.New("ErrWindow window behind max")
	ErrWindowABMin     = errors.New("ErrWindow buffer ahead min")
	ErrWindowABMax     = errors.New("ErrWindow buffer ahead max")
	ErrWindowBBMin     = errors.New("ErrWindow buffer behind min")
	ErrWindowBBMax     = errors.New("ErrWindow buffer behind max")
	ErrWindowDegreeMin = errors.New("ErrWindow degree min")
	ErrWindowDegreeMax = errors.New("ErrWindow degree max")
)

// validateNew performs simple min/max checks of the Tracker creation variables
func validateNew(aw uint16, bw uint16, ab uint16, bb uint16, degree int) error {

	if aw <= MinWindowCst {
		log.Printf("aw:%v, aw < MinWindowCst:%v", aw, MinWindowCst)
		return ErrWindowAWMin
	}
	if aw > MaxWindowCst {
		log.Printf("aw:%v, aw > MaxWindowCst:%v", aw, MaxWindowCst)
		return ErrWindowAWMax
	}
	if bw <= MinWindowCst {
		log.Printf("bw:%v, bw < MinWindowCst:%v", bw, MinWindowCst)
		return ErrWindowBWMin
	}
	if bw > MaxWindowCst {
		log.Printf("bw:%v, bw > MaxWindowCst:%v", bw, MaxWindowCst)
		return ErrWindowBWMax
	}

	if ab <= MinWindowCst {
		log.Printf("ab:%v, ab < MinWindowCst:%v", ab, MinWindowCst)
		return ErrWindowABMin
	}
	if ab > MaxWindowCst {
		log.Printf("ab:%v, ab > MaxWindowCst:%v", ab, MaxWindowCst)
		return ErrWindowABMax
	}
	if bb <= MinWindowCst {
		log.Printf("bb:%v, bb < MinWindowCst:%v", bb, MinWindowCst)
		return ErrWindowBBMin
	}
	if bb > MaxWindowCst {
		log.Printf("bb:%v, bb > MaxWindowCst:%v", bb, MaxWindowCst)
		return ErrWindowBBMax
	}

	if degree < MinDegree {
		log.Printf("degree:%v, degree < MinDegree:%v", degree, MinDegree)
		return ErrWindowDegreeMin
	}
	if degree > MaxDegree {
		log.Printf("degree:%v, degree > MaxDegree:%v", degree, MaxDegree)
		return ErrWindowDegreeMax
	}

	return nil
}
