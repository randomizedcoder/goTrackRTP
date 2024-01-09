all: test

test:
	go test -v

i:
	go test -v -run TestTrackerInit

tw:
	go test -v -run TestTrackerWindow

long:
	LONG=true go test -v -run TestLongRunningLoop

lw:
	LONG=true go test -v -run TestLongRunningWindow
#-test.timeout=30m

db:
	LONG=true go test -v -run TestLongRunningBackwardDuplicates

ss:
	LONG=true go test -v -run TestLongRunningSkipSend

# math tests
math: l d

l:
	go test -v -run TestIsLess
	go test -v -run TestIsLessBranchless
	go test -v -run TestIsLessBranch

d:
	go test -v -run TestDiff