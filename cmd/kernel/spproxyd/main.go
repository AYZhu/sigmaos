package main

import (
	"os"

	db "sigmaos/debug"
	"sigmaos/spproxy"
)

func main() {
	if len(os.Args) != 1 {
		db.DFatalf("Usage: %v", os.Args[0])
	}
	if err := spproxy.RunSPProxySrv(true); err != nil {
		db.DFatalf("Fatal start: %v %v\n", os.Args[0], err)
	}
}
