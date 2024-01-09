package main

// Nasty enum code generator for trackingRing
// Honestly, it might be better to use protobufs....

import (
	"fmt"
	"log"
)

var (
	position    = []string{"Unknown", "Ahead", "Behind"}
	category    = []string{"Unknown", "Window", "Buffer", "Reset"}
	subcategory = []string{"Unknown", "Next", "Duplicate"}
)

func main() {

	log.Println("enum_gen")

	fmt.Printf("const (\n")

	for i, p := range position {
		for j, c := range category {
			for k, s := range subcategory {
				fmt.Printf("\t%s%s%s", p, c, s)
				if i == 0 && j == 0 && k == 0 {
					fmt.Printf(" int = iota")
				}
				fmt.Printf("\n")
			}
		}
	}

	fmt.Printf(")\n")
}
