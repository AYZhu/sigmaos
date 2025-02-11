package main

import (
	"errors"
	"os"
	"path"
	"runtime"

	db "sigmaos/debug"
	"sigmaos/fsetcd"
	"sigmaos/proc"
	"sigmaos/sigmaclnt"
	sp "sigmaos/sigmap"
)

func main() {
	if len(os.Args) < 2 {
		db.DFatalf("Usage: %v out\n", os.Args[0])
	}
	l, err := MakeSpinner(os.Args[1:])
	if err != nil {
		db.DFatalf("%v: error %v", os.Args[0], err)
	}
	l.Work()
}

type Spinner struct {
	*sigmaclnt.SigmaClnt
	outdir string
}

func MakeSpinner(args []string) (*Spinner, error) {
	if len(args) < 1 {
		return nil, errors.New("MakeSpinner: too few arguments")
	}
	s := &Spinner{}
	sc, err := sigmaclnt.MkSigmaClnt("spinner")
	if err != nil {
		return nil, err
	}
	s.SigmaClnt = sc
	s.outdir = args[0]

	db.DPrintf(db.SPINNER, "MakeSpinner: %v\n", args)

	li, err := sc.LeaseClnt.AskLease(s.outdir, fsetcd.LeaseTTL)
	if err != nil {
		return nil, err
	}
	li.KeepExtending()

	if _, err := s.PutFileEphemeral(path.Join(s.outdir, proc.GetPid().String()), 0777, sp.OWRITE, li.Lease(), []byte{}); err != nil {
		db.DFatalf("MakeFile error: %v", err)
	}

	err = s.Started()
	if err != nil {
		db.DFatalf("Started: error %v\n", err)
	}
	return s, nil
}

func (s *Spinner) waitEvict() {
	err := s.WaitEvict(proc.GetPid())
	if err != nil {
		db.DFatalf("Error WaitEvict: %v", err)
	}
	s.ClntExit(proc.MakeStatus(proc.StatusEvicted))
	os.Exit(0)
}

func (s *Spinner) spin() {
	for {
		runtime.Gosched()
	}
}

func (s *Spinner) Work() {
	go s.spin()
	s.waitEvict()
}
