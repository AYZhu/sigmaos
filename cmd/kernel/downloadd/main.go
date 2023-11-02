package main

import (
	"os"

	db "sigmaos/debug"
	"sigmaos/downloadd"
)

func main() {
	if len(os.Args) != 2 {
		db.DFatalf("Usage: %v kernelId", os.Args[0])
	}
	if err := downloadd.RunDownloadd(os.Args[1]); err != nil {
		db.DFatalf("Fatal start: %v %v\n", os.Args[0], err)
	}
}
