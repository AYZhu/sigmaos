package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	db "sigmaos/debug"
	"sigmaos/fs"
	"sigmaos/pipe"
	"sigmaos/proc"
	"sigmaos/sigmaclnt"
	sp "sigmaos/sigmap"
)

//
// Reads data from input (os.Args[2])), and writes it to the named pipe found
// at proc.PARENTDIR/proc.SHARED
//

func main() {
	r, err := MakeReader(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: error %v", os.Args[0], err)
		os.Exit(1)
	}
	s := r.Work()
	r.Exit(s)
}

type Reader struct {
	*sigmaclnt.SigmaClnt
	input  string
	output string
	ctx    fs.CtxI
}

func MakeReader(args []string) (*Reader, error) {
	if len(args) != 2 {
		return nil, errors.New("MakeReader: too few arguments")
	}
	log.Printf("MakeReader %v: %v\n", proc.GetPid(), args)
	r := &Reader{}
	sc, err := sigmaclnt.NewSigmaClnt("fsreader")
	if err != nil {
		return nil, err
	}
	r.SigmaClnt = sc
	r.input = args[1]
	r.output = path.Join(proc.PARENTDIR, proc.SHARED) + "/"
	r.Started()
	return r, nil
}

func (r *Reader) Work() *proc.Status {
	db.DPrintf(db.FSREADER, "Reader: %v\n", r.input)
	// Open the pipe.
	pipefd, err := r.Open(r.output, sp.OWRITE)
	if err != nil {
		db.DFatalf("%v: Open error: %v", proc.GetProgram(), err)
	}
	defer r.Close(pipefd)
	fd, err := r.Open(r.input, sp.OREAD)
	if err != nil {
		return proc.MakeStatusErr("File not found", nil)
	}
	defer r.Close(fd)
	for {
		data, err := r.Read(fd, pipe.PIPESZ)
		if len(data) == 0 || err != nil {
			break
		}
		_, err = r.Write(pipefd, data)
		if err != nil {
			db.DFatalf("%v: Error pipe Write: %v", proc.GetProgram(), err)
		}
	}
	return proc.MakeStatus(proc.StatusOK)
}

func (r *Reader) Exit(status *proc.Status) {
	r.Exit(status)
}
