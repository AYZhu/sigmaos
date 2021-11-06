package main

import (
	"fmt"
	"os"

	"ulambda/twopc"
)

func main() {
	cd, err := twopc.MakeCoord(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: error %v", os.Args[0], err)
		os.Exit(1)
	}
	cd.Exit()
	cd.TwoPC()
}
