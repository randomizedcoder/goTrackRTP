package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"syscall"

	"github.com/randomizedcoder/goTrackRTP"

	_ "unsafe"
)

const (
	loopsCst = math.MaxInt32

	debugLevelCst = 11

	signalChannelSize = 10

	awCst = 100
	bwCst = 100
	abCst = 100
	bbCst = 100

	// maxRandJumpCst = 10
)

var (
	// Passed by "go build -ldflags" for the show version
	commit string
	date   string
)

func main() {

	log.Println("goTrackingRTPer")

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	go initSignalHandler(cancel)

	loops := flag.Int("loops", loopsCst, "loops")

	version := flag.Bool("version", false, "version")

	aw := flag.Int("aw", awCst, "ahead window")
	bw := flag.Int("bw", bwCst, "behind window")
	ab := flag.Int("ab", abCst, "ahead buffer")
	bb := flag.Int("bb", bbCst, "behind buffer")

	dl := flag.Int("dl", debugLevelCst, "nasty debugLevel")

	flag.Parse()

	if *version {
		fmt.Println("commit:", commit, "\tdate(UTC):", date)
		os.Exit(0)
	}

	tr, err := goTrackRTP.New(uint16(*aw), uint16(*bw), uint16(*ab), uint16(*bb), *dl)
	if err != nil {
		log.Fatal("goTrackRTP.New:", err)
	}

	fmt.Println("loops:", loops)

	var s uint16
	for i := 0; i < *loops; i++ {
		tax, err := tr.PacketArrival(s)
		if err != nil {
			log.Fatal("PacketArrival:", err)
		}
		if i > *aw+*bw && tax.Len < 100 {
			log.Fatalf("tax.Len:%d should be < 100:", tax.Len)
		}
		s++
	}

	log.Println("go_generate_rtp: That's all Folks!")
}

// initSignalHandler sets up signal handling for the process, and
// will call cancel() when recieved
func initSignalHandler(cancel context.CancelFunc) {
	c := make(chan os.Signal, signalChannelSize)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Printf("Signal caught, closing application")
	cancel()
	os.Exit(0)
}

//go:linkname FastRand runtime.fastrand
func FastRand() uint32

// https://cs.opensource.google/go/go/+/master:src/runtime/stubs.go;l=151?q=FastRandN&ss=go%2Fgo
// https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/

//go:linkname FastRandN runtime.fastrandn
func FastRandN(n uint32) uint32
