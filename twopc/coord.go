package twopc

//
// Coordinator for two-phase commit.  This is a short-lived daemon: it
// performs the transaction and then exits.
//

import (
	"fmt"
	"log"
	"os"

	db "sigmaos/debug"
	//	"sigmaos/fenceclnt"
	"sigmaos/fslib"
	"sigmaos/proc"
	"sigmaos/procclnt"
	sp "sigmaos/sigmap"
)

const (
	DIR2PC         = "name/twopc"
	COORD          = DIR2PC + "/coord"
	TWOPC          = DIR2PC + "/twopc"
	TWOPCPREP      = DIR2PC + "/twopcprep"
	TWOPCCOMMIT    = DIR2PC + "/twopccommit"
	TWOPCPREPARED  = DIR2PC + "/prepared/"
	TWOPCCOMMITTED = DIR2PC + "/committed/"
	TWOPCFENCE     = DIR2PC + "/fence"
)

type Coord struct {
	*fslib.FsLib
	*procclnt.ProcClnt
	opcode string
	args   []string
	ch     chan Tstatus
	twopc  *Twopc
	//	fclnt  *fenceclnt.FenceClnt
}

func MakeCoord(args []string) (*Coord, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("MakeCoord: too few arguments %v\n", args)
	}
	cd := &Coord{}
	cd.opcode = args[0]
	cd.args = args[1:]
	cd.ch = make(chan Tstatus)
	cd.FsLib = fslib.MakeFsLib("coord")
	cd.ProcClnt = procclnt.MakeProcClnt(cd.FsLib)
	//	cd.fclnt = fenceclnt.MakeFenceClnt(cd.FsLib, TWOPCFENCE, 0, []string{DIR2PC})

	// Grab fence before starting coord
	//	cd.fclnt.AcquireFenceW([]byte{})

	log.Printf("COORD lock %v\n", args)

	db.DPrintf("COORD", "New coord %v", args)

	if _, err := cd.PutFile(COORD, 0777|sp.DMTMP, sp.OWRITE, nil); err != nil {
		db.DFatalf("MakeFile %v failed %v\n", COORD, err)
	}

	cd.Started()
	return cd, nil
}

func (cd *Coord) exit() {
	log.Printf("unlock\n")

	if err := cd.Remove(COORD); err != nil {
		log.Printf("Remove %v failed %v\n", COORD, err)
	}

	//	cd.fclnt.ReleaseFence()
}

func (cd *Coord) restart() {
	cd.twopc = clean(cd.FsLib)
	if cd.twopc == nil {
		log.Printf("COORD clean\n")
		return
	}
	prepared := mkFlwsMapStatus(cd.FsLib, TWOPCPREPARED)
	committed := mkFlwsMapStatus(cd.FsLib, TWOPCCOMMITTED)

	db.DPrintf("COORD", "Restart: twopc %v prepared %v commit %v\n",
		cd.twopc, prepared, committed)

	fws := mkFlwsMap(cd.FsLib, cd.twopc.Participants)
	if fws.doCommit(prepared) {

		if committed.len() == fws.len() {
			db.DPrintf("COORD", "Restart: finished commit %d\n", committed.len())
			cd.cleanup()
		} else {
			db.DPrintf("COORD", "Restart: finish commit %d\n", committed.len())
			cd.commit(fws, committed.len(), true)
		}
	} else {
		db.DPrintf("COORD", "Restart: abort\n")
		cd.commit(fws, committed.len(), false)
	}
}

func (cd *Coord) rmStatusFiles(dir string) {
	sts, err := cd.GetDir(dir)
	if err != nil {
		db.DFatalf("COORD: ReadDir commit error %v\n", err)
	}
	for _, st := range sts {
		fn := dir + st.Name
		err = cd.Remove(fn)
		if err != nil {
			db.DPrintf("COORD", "Remove %v failed %v\n", fn, err)
		}
	}
}

func (cd *Coord) watchStatus(p string, err error) {
	db.DPrintf("COORD", "watchStatus %v\n", p)
	status := TABORT
	b, err := cd.GetFile(p)
	if err != nil {
		db.DPrintf("COORD", "watchStatus ReadFile %v err %v\n", p, b)
	}
	if string(b) == "OK" {
		status = TCOMMIT
	}
	cd.ch <- status
}

func (cd *Coord) watchFlw(p string, err error) {
	db.DPrintf("COORD", "watchFlw %v\n", p)
	cd.ch <- TCRASH
}

func (cd *Coord) prepare(nextFws *FlwsMap) (bool, int) {
	nextFws.setStatusWatches(TWOPCPREPARED, cd.watchStatus)

	err := cd.PutFileJsonAtomic(TWOPCPREP, 0777, *cd.twopc)
	if err != nil {
		db.DPrintf("COORD", "COORD: MakeFileJsonAtomic %v err %v\n",
			TWOPCCOMMIT, err)
	}

	// depending how many KVs ack, crash3 results
	// in a abort or commit
	if cd.opcode == "crash3" {
		db.DPrintf("COORD", "Crash3\n")
		os.Exit(1)
	}

	success := true
	n := 0
	for i := 0; i < nextFws.len(); i++ {
		status := <-cd.ch
		switch status {
		case TCOMMIT:
			db.DPrintf("COORD", "KV prepared\n")
			n += 1
		case TABORT:
			db.DPrintf("COORD", "KV aborted\n")
			n += 1
			success = false
		default:
			db.DPrintf("COORD", "KV crashed\n")
			success = false
		}
	}
	return success, n
}

func (cd *Coord) commit(fws *FlwsMap, ndone int, ok bool) {
	if ok {
		cd.twopc.Status = TCOMMIT
		db.DPrintf("COORD", "Commit to %v\n", cd.twopc)
	} else {
		cd.twopc.Status = TABORT
		db.DPrintf("COORD", "Abort to %v\n", cd.twopc)
	}

	if err := cd.SetFileJson(TWOPCPREP, *cd.twopc); err != nil {
		db.DPrintf("COORD", "Write %v err %v\n", TWOPCPREP, err)
		return
	}

	fws.setStatusWatches(TWOPCCOMMITTED, cd.watchStatus)

	// commit/abort to new TWOPC, which maybe the same as the
	// old one
	err := cd.Rename(TWOPCPREP, TWOPCCOMMIT)
	if err != nil {
		db.DPrintf("COORD", "COORD: rename %v -> %v: error %v\n",
			TWOPCPREP, TWOPCCOMMIT, err)
		return
	}

	// crash4 should results in commit (assuming no KVs crash)
	if cd.opcode == "crash4" {
		db.DPrintf("COORD", "Crash4\n")
		os.Exit(1)
	}

	for i := 0; i < fws.len()-ndone; i++ {
		s := <-cd.ch
		db.DPrintf("COORD", "KV commit status %v\n", s)
	}

	db.DPrintf("COORD", "Done commit/abort\n")

	cd.cleanup()
}

func (cd *Coord) TwoPC() {
	defer cd.exit()

	log.Printf("COORD Coord: %v\n", cd.args)

	db.DPrintf("COORD", "Coord: %v\n", cd.args)

	// XXX set removeWatch on KVs? maybe in KV

	cd.restart()

	switch cd.opcode {
	case "restart":
		return
	}

	cd.twopc = makeTwopc(1, cd.args)

	fws := mkFlwsMap(cd.FsLib, cd.args)

	db.DPrintf("COORD", "Coord twopc %v %v\n", cd.twopc, fws)

	if cd.opcode == "crash2" {
		log.Printf("crash2\n")
		db.DPrintf("COORD", "Crash2\n")
		os.Exit(1)
	}

	cd.Remove(TWOPCCOMMIT) // don't care if succeeds or not
	cd.rmStatusFiles(TWOPCPREPARED)
	cd.rmStatusFiles(TWOPCCOMMITTED)

	fws.setFlwsWatches(cd.watchFlw)

	log.Printf("COORD prepare\n")

	ok, n := cd.prepare(fws)

	log.Printf("COORD commit %v\n", ok)

	cd.commit(fws, fws.len()-n, ok)
}

func (cd *Coord) cleanup() {
	log.Printf("COORD cleanup %v\n", TWOPCCOMMIT)
	cd.Remove(TWOPCCOMMIT) // don't care if succeeds or not
}

func (cd *Coord) Exit() {
	cd.Exited(proc.MakeStatus(proc.StatusOK))
}
