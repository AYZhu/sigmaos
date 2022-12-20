package protsrv

import (
	"sigmaos/fid"
	"sigmaos/fs"
	"sigmaos/lockmap"
	"sigmaos/namei"
	"sigmaos/path"
	"sigmaos/serr"
)

// LookupObj/namei will return an lo and a locked watch for it, even
// in error cases because the caller create a new fid anyway.
func (ps *ProtSrv) lookupObj(ctx fs.CtxI, po *fid.Pobj, target path.Path) ([]fs.FsObj, fs.FsObj, *lockmap.PathLock, path.Path, *serr.Err) {
	src := po.Path()
	lk := ps.plt.Acquire(ctx, src)
	o := po.Obj()
	if len(target) == 0 {
		ps.stats.IncPath(src)
		return nil, o, lk, nil, nil
	}
	if !o.Perm().IsDir() {
		ps.stats.IncPath(src)
		return nil, o, lk, nil, serr.MkErr(serr.TErrNotDir, src.Base())
	}
	ps.stats.IncPathString(lk.Path())
	return namei.Walk(ps.plt, ctx, o, lk, src, target, nil)
}
