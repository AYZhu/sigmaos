package fsux

import (
	"sync"

	"sigmaos/container"
	db "sigmaos/debug"
	"sigmaos/proc"
	"sigmaos/sigmaclnt"
	sp "sigmaos/sigmap"
	"sigmaos/sigmasrv"
	// "sigmaos/seccomp"
)

var fsux *FsUx

type FsUx struct {
	*sigmaclnt.SigmaClnt
	*sigmasrv.SigmaSrv
	mount string

	sync.Mutex
	ot *ObjTable
}

func RunFsUx(rootux string) {
	ip, err := container.LocalIP()
	if err != nil {
		db.DFatalf("LocalIP %v %v\n", sp.UX, err)
	}
	// seccomp.LoadFilter()  // sanity check: if enabled we want fsux to fail
	fsux := newUx(rootux)
	root, sr := makeDir([]string{rootux})
	if sr != nil {
		db.DFatalf("%v: makeDir %v\n", proc.GetName(), sr)
	}
	srv, err := sigmasrv.MakeSigmaSrvRoot(root, ip+":0", sp.UX, sp.UXREL)
	if err != nil {
		db.DFatalf("%v: BootSrvAndPost %v\n", proc.GetName(), err)
	}
	fsux.SigmaSrv = srv
	fsux.RunServer()
}

func newUx(rootux string) *FsUx {
	fsux = &FsUx{}
	fsux.ot = MkObjTable()
	return fsux
}
