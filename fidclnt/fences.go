package fidclnt

import (
	"sync"

	db "sigmaos/debug"
	"sigmaos/path"
	"sigmaos/serr"
	sp "sigmaos/sigmap"
)

//
// Keep track of registered fences for a pathname so that caller of
// pathclnt doesn't have to provide a fence as argument on each call.
//

type FenceTable struct {
	sync.Mutex
	fencedDirs map[string]sp.Tfence
}

func MakeFenceTable() *FenceTable {
	ft := &FenceTable{}
	ft.fencedDirs = make(map[string]sp.Tfence)
	return ft
}

// If already exist, just update
func (ft *FenceTable) Insert(p string, f sp.Tfence) *serr.Err {
	ft.Lock()
	defer ft.Unlock()

	path := path.Split(p) // cleans up p

	db.DPrintf(db.FIDCLNT, "Insert fence %v %v\n", path, f)
	ft.fencedDirs[path.String()] = f
	return nil
}

func (ft *FenceTable) Lookup(p path.Path) *sp.Tfence {
	ft.Lock()
	defer ft.Unlock()

	for pn, f := range ft.fencedDirs {
		db.DPrintf(db.FIDCLNT, "Lookup fence %v %v\n", p, f)
		if p.IsParent(path.Split(pn)) {
			return &f
		}
	}
	db.DPrintf(db.FIDCLNT, "Lookup fence %v: no fence\n", p)
	return sp.NullFence()
}
