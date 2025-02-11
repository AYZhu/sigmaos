package main

import (
	"os"

	db "sigmaos/debug"
	"sigmaos/kv"
)

func main() {
	if len(os.Args) != 7 {
		db.DFatalf("%v: <job> <epoch> <shard> <src> <dst> <repl>\n", os.Args[0])
	}
	mv, err := kv.MakeMover(os.Args[1], os.Args[2], os.Args[3], os.Args[4], os.Args[5], os.Args[6])
	if err != nil {
		db.DFatalf("Error MakeMover: %v", err)
	}
	mv.Move(os.Args[4], os.Args[5])
}
