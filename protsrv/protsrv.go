package protsrv

import (
	db "sigmaos/debug"
	"sigmaos/fcall"
	"sigmaos/fid"
	"sigmaos/fs"
	"sigmaos/lockmap"
	"sigmaos/namei"
	"sigmaos/path"
	"sigmaos/sesssrv"
	np "sigmaos/sigmap"
	"sigmaos/stats"
	"sigmaos/version"
	"sigmaos/watch"
)

//
// There is one protsrv per session, but they share the watch table,
// version table, and stats across sessions.  Each session has its own
// fid table, ephemeral table, and lease table.
//

type ProtSrv struct {
	ssrv  *sesssrv.SessSrv
	plt   *lockmap.PathLockTable // shared across sessions
	wt    *watch.WatchTable      // shared across sessions
	vt    *version.VersionTable  // shared across sessions
	ft    *fidTable
	et    *ephemeralTable
	stats *stats.Stats
	sid   fcall.Tsession
}

func MakeProtServer(s np.SessServer, sid fcall.Tsession) np.Protsrv {
	ps := &ProtSrv{}
	srv := s.(*sesssrv.SessSrv)
	ps.ssrv = srv

	ps.ft = makeFidTable()
	ps.et = makeEphemeralTable()
	ps.plt = srv.GetPathLockTable()
	ps.wt = srv.GetWatchTable()
	ps.vt = srv.GetVersionTable()
	ps.stats = srv.GetStats()
	ps.sid = sid
	db.DPrintf("PROTSRV", "MakeProtSrv -> %v", ps)
	return ps
}

func (ps *ProtSrv) mkQid(perm np.Tperm, path np.Tpath) *np.Tqid {
	return np.MakeQidPerm(perm, ps.vt.GetVersion(path), path)
}

func (ps *ProtSrv) Version(args *np.Tversion, rets *np.Rversion) *np.Rerror {
	rets.Msize = args.Msize
	rets.Version = "9P2000"
	return nil
}

func (ps *ProtSrv) Auth(args *np.Tauth, rets *np.Rauth) *np.Rerror {
	return np.MkRerror(fcall.MkErr(fcall.TErrNotSupported, "Auth"))
}

func (ps *ProtSrv) Attach(args *np.Tattach, rets *np.Rattach) *np.Rerror {
	db.DPrintf("PROTSRV", "Attach %v", args)
	p := path.Split(args.Aname)
	root, ctx := ps.ssrv.AttachTree(args.Uname, args.Aname, ps.sid)
	tree := root.(fs.FsObj)
	qid := ps.mkQid(tree.Perm(), tree.Path())
	if args.Aname != "" {
		dlk := ps.plt.Acquire(ctx, path.Path{})
		_, lo, lk, rest, err := namei.Walk(ps.plt, ctx, root, dlk, path.Path{}, p, nil)
		defer ps.plt.Release(ctx, lk)
		if len(rest) > 0 || err != nil {
			return np.MkRerror(err)
		}
		// insert before releasing
		ps.vt.Insert(lo.Path())
		tree = lo
		qid = ps.mkQid(lo.Perm(), lo.Path())
	} else {
		// root is already in the version table; this updates
		// just the refcnt.
		ps.vt.Insert(root.Path())
	}
	ps.ft.Add(args.Fid, fid.MakeFidPath(fid.MkPobj(p, tree, ctx), 0, qid))
	rets.Qid = qid
	return nil
}

// Delete ephemeral files created on this session.
func (ps *ProtSrv) Detach(rets *np.Rdetach, detach np.DetachF) *np.Rerror {
	db.DPrintf("PROTSRV", "Detach %v eph %v", ps.sid, ps.et.Get())

	// Several threads maybe waiting in a sesscond. DeleteSess
	// will unblock them so that they can bail out.
	ps.ssrv.GetSessCondTable().DeleteSess(ps.sid)

	ps.ft.ClunkOpen()
	ephemeral := ps.et.Get()
	for _, po := range ephemeral {
		db.DPrintf("PROTSRV", "Detach %v", po.Path())
		ps.removeObj(po.Ctx(), po.Obj(), po.Path())
	}
	if detach != nil {
		detach(ps.sid)
	}
	return nil
}

func (ps *ProtSrv) makeQids(os []fs.FsObj) []*np.Tqid {
	var qids []*np.Tqid
	for _, o := range os {
		qids = append(qids, ps.mkQid(o.Perm(), o.Path()))
	}
	return qids
}

func (ps *ProtSrv) lookupObjLast(ctx fs.CtxI, f *fid.Fid, names path.Path, resolve bool) (fs.FsObj, *fcall.Err) {
	_, lo, lk, _, err := ps.lookupObj(ctx, f.Pobj(), names)
	ps.plt.Release(ctx, lk)
	if err != nil {
		return nil, err
	}
	if lo.Perm().IsSymlink() && resolve {
		return nil, fcall.MkErr(fcall.TErrNotDir, names[len(names)-1])
	}
	return lo, nil
}

// Requests that combine walk, open, and do operation in a single RPC,
// which also avoids clunking. They may fail because args.Wnames may
// contains a special path element; in that, case the client must walk
// args.Wnames.
func (ps *ProtSrv) Walk(args *np.Twalk, rets *np.Rwalk) *np.Rerror {
	f, err := ps.ft.Lookup(args.Tfid())
	if err != nil {
		return np.MkRerror(err)
	}

	db.DPrintf("PROTSRV", "%v: Walk o %v args %v (%v)", f.Pobj().Ctx().Uname(), f, args, len(args.Wnames))

	os, lo, lk, rest, err := ps.lookupObj(f.Pobj().Ctx(), f.Pobj(), args.Wnames)
	defer ps.plt.Release(f.Pobj().Ctx(), lk)
	if err != nil && !fcall.IsMaybeSpecialElem(err) {
		return np.MkRerror(err)
	}

	// let the client decide what to do with rest (when there is a rest)
	n := len(args.Wnames) - len(rest)
	p := append(f.Pobj().Path().Copy(), args.Wnames[:n]...)
	rets.Qids = ps.makeQids(os)
	qid := ps.mkQid(lo.Perm(), lo.Path())
	db.DPrintf("PROTSRV", "%v: Walk MakeFidPath fid %v p %v lo %v qid %v os %v", args.NewFid, f.Pobj().Ctx().Uname(), p, lo, qid, os)
	ps.ft.Add(args.Tnewfid(), fid.MakeFidPath(fid.MkPobj(p, lo, f.Pobj().Ctx()), 0, qid))

	ps.vt.Insert(qid.Tpath())

	return nil
}

func (ps *ProtSrv) Clunk(args *np.Tclunk, rets *np.Rclunk) *np.Rerror {
	f, err := ps.ft.Lookup(args.Tfid())
	if err != nil {
		return np.MkRerror(err)
	}
	db.DPrintf("PROTSRV", "%v: Clunk %v f %v path %v", f.Pobj().Ctx().Uname(), args.Fid, f, f.Pobj().Path())
	if f.IsOpen() { // has the fid been opened?
		f.Pobj().Obj().Close(f.Pobj().Ctx(), f.Mode())
		f.Close()
	}
	ps.ft.Del(args.Tfid())
	ps.vt.Delete(f.Pobj().Obj().Path())
	return nil
}

func (ps *ProtSrv) Open(args *np.Topen, rets *np.Ropen) *np.Rerror {
	f, err := ps.ft.Lookup(args.Fid)
	if err != nil {
		return np.MkRerror(err)
	}
	db.DPrintf("PROTSRV", "%v: Open f %v %v", f.Pobj().Ctx().Uname(), f, args)

	o := f.Pobj().Obj()
	no, r := o.Open(f.Pobj().Ctx(), args.Mode)
	if r != nil {
		return np.MkRerror(r)
	}
	f.SetMode(args.Mode)
	if no != nil {
		f.Pobj().SetObj(no)
		ps.vt.Insert(no.Path())
		ps.vt.IncVersion(no.Path())
		rets.Qid = *ps.mkQid(no.Perm(), no.Path())
	} else {
		rets.Qid = *ps.mkQid(o.Perm(), o.Path())
	}
	return nil
}

func (ps *ProtSrv) Watch(args *np.Twatch, rets *np.Ropen) *np.Rerror {
	f, err := ps.ft.Lookup(args.Fid)
	if err != nil {
		return np.MkRerror(err)
	}
	p := f.Pobj().Path()
	ino := f.Pobj().Obj().Path()

	db.DPrintf("PROTSRV", "%v: Watch %v v %v %v", f.Pobj().Ctx().Uname(), f.Pobj().Path(), f.Qid(), args)

	// get path lock on for p, so that remove cannot remove file
	// before watch is set.
	pl := ps.plt.Acquire(f.Pobj().Ctx(), p)
	defer ps.plt.Release(f.Pobj().Ctx(), pl)

	v := ps.vt.GetVersion(ino)
	if !np.VEq(f.Qid().Tversion(), v) {
		return np.MkRerror(fcall.MkErr(fcall.TErrVersion, v))
	}
	err = ps.wt.WaitWatch(pl, ps.sid)
	if err != nil {
		return np.MkRerror(err)
	}
	return nil
}

func (ps *ProtSrv) makeFid(ctx fs.CtxI, dir path.Path, name string, o fs.FsObj, eph bool, qid *np.Tqid) *fid.Fid {
	p := dir.Copy()
	po := fid.MkPobj(append(p, name), o, ctx)
	nf := fid.MakeFidPath(po, 0, qid)
	if eph {
		ps.et.Add(o, po)
	}
	return nf
}

// Create name in dir. If OWATCH is set and name already exits, wait
// until another thread deletes it, and retry.
func (ps *ProtSrv) createObj(ctx fs.CtxI, d fs.Dir, dlk *lockmap.PathLock, fn path.Path, perm np.Tperm, mode np.Tmode) (fs.FsObj, *lockmap.PathLock, *fcall.Err) {
	name := fn.Base()
	if name == "." {
		return nil, nil, fcall.MkErr(fcall.TErrInval, name)
	}
	for {
		flk := ps.plt.Acquire(ctx, fn)
		o1, err := d.Create(ctx, name, perm, mode)
		db.DPrintf("PROTSRV", "%v: Create %v %v %v ephemeral %v %v", ctx.Uname(), name, o1, err, perm.IsEphemeral(), ps.sid)
		if err == nil {
			ps.wt.WakeupWatch(dlk)
			return o1, flk, nil
		} else {
			ps.plt.Release(ctx, flk)
			if mode&np.OWATCH == np.OWATCH && err.Code() == fcall.TErrExists {
				err := ps.wt.WaitWatch(dlk, ps.sid)
				db.DPrintf("PROTSRV", "%v: Create: Wait %v %v sid %v err %v", ctx.Uname(), name, o1, ps.sid, err)
				if err != nil {
					return nil, nil, err
				}
				// try again; we will hold lock on watchers
			} else {
				return nil, nil, err
			}
		}
	}
}

func (ps *ProtSrv) Create(args *np.Tcreate, rets *np.Rcreate) *np.Rerror {
	f, err := ps.ft.Lookup(args.Fid)
	if err != nil {
		return np.MkRerror(err)
	}
	db.DPrintf("PROTSRV", "%v: Create f %v", f.Pobj().Ctx().Uname(), f)
	o := f.Pobj().Obj()
	fn := f.Pobj().Path().Append(args.Name)
	if !o.Perm().IsDir() {
		return np.MkRerror(fcall.MkErr(fcall.TErrNotDir, f.Pobj().Path()))
	}
	d := o.(fs.Dir)
	dlk := ps.plt.Acquire(f.Pobj().Ctx(), f.Pobj().Path())
	defer ps.plt.Release(f.Pobj().Ctx(), dlk)

	o1, flk, err := ps.createObj(f.Pobj().Ctx(), d, dlk, fn, args.Perm, args.Mode)
	if err != nil {
		return np.MkRerror(err)
	}
	defer ps.plt.Release(f.Pobj().Ctx(), flk)
	ps.vt.Insert(o1.Path())
	ps.vt.IncVersion(o1.Path())
	qid := ps.mkQid(o1.Perm(), o1.Path())
	nf := ps.makeFid(f.Pobj().Ctx(), f.Pobj().Path(), args.Name, o1, args.Perm.IsEphemeral(), qid)
	ps.ft.Add(args.Fid, nf)
	ps.vt.IncVersion(f.Pobj().Obj().Path())
	nf.SetMode(args.Mode)
	rets.Qid = *qid
	return nil
}

func (ps *ProtSrv) Flush(args *np.Tflush, rets *np.Rflush) *np.Rerror {
	return nil
}

func (ps *ProtSrv) Read(args *np.Tread, rets *np.Rread) *np.Rerror {
	f, err := ps.ft.Lookup(args.Fid)
	if err != nil {
		return np.MkRerror(err)
	}
	db.DPrintf("PROTSRV", "%v: Read f %v args %v", f.Pobj().Ctx().Uname(), f, args)
	err = f.Read(args.Offset, args.Count, np.NoV, rets)
	if err != nil {
		return np.MkRerror(err)
	}
	return nil
}

func (ps *ProtSrv) ReadV(args *np.TreadV, rets *np.Rread) *np.Rerror {
	f, err := ps.ft.Lookup(args.Tfid())
	if err != nil {
		return np.MkRerror(err)
	}
	v := ps.vt.GetVersion(f.Pobj().Obj().Path())
	db.DPrintf("PROTSRV1", "%v: ReadV f %v args %v v %d", f.Pobj().Ctx().Uname(), f, args, v)
	if !np.VEq(args.Tversion(), v) {
		return np.MkRerror(fcall.MkErr(fcall.TErrVersion, v))
	}

	err = f.Read(args.Toffset(), args.Tcount(), args.Tversion(), rets)
	if err != nil {
		return np.MkRerror(err)
	}
	return nil
}

func (ps *ProtSrv) Write(args *np.Twrite, rets *np.Rwrite) *np.Rerror {
	f, err := ps.ft.Lookup(args.Fid)
	if err != nil {
		return np.MkRerror(err)
	}
	n, err := f.Write(args.Offset, args.Data, np.NoV)
	if err != nil {
		return np.MkRerror(err)
	}
	rets.Count = uint32(n)
	ps.vt.IncVersion(f.Pobj().Obj().Path())
	return nil
}

func (ps *ProtSrv) WriteRead(args *np.Twriteread, rets *np.Rread) *np.Rerror {
	f, err := ps.ft.Lookup(np.Tfid(args.Fid))
	if err != nil {
		return np.MkRerror(err)
	}
	db.DPrintf("PROTSRV0", "%v: WriteRead %v args %v path %d\n", f.Pobj().Ctx().Uname(), f.Pobj().Path(), args, f.Pobj().Obj().Path())
	rets.Data, err = f.WriteRead(args.Data)
	if err != nil {
		return np.MkRerror(err)
	}
	ps.vt.IncVersion(f.Pobj().Obj().Path())
	return nil
}

func (ps *ProtSrv) WriteV(args *np.TwriteV, rets *np.Rwrite) *np.Rerror {
	f, err := ps.ft.Lookup(args.Tfid())
	if err != nil {
		return np.MkRerror(err)
	}
	v := ps.vt.GetVersion(f.Pobj().Obj().Path())
	db.DPrintf("PROTSRV1", "%v: WriteV %v args %v path %d v %d", f.Pobj().Ctx().Uname(), f.Pobj().Path(), args, f.Pobj().Obj().Path(), v)
	if !np.VEq(args.Tversion(), v) {
		return np.MkRerror(fcall.MkErr(fcall.TErrVersion, v))
	}
	n, err := f.Write(args.Toffset(), args.Data, args.Tversion())
	if err != nil {
		return np.MkRerror(err)
	}
	rets.Count = uint32(n)
	ps.vt.IncVersion(f.Pobj().Obj().Path())
	return nil
}

func (ps *ProtSrv) removeObj(ctx fs.CtxI, o fs.FsObj, path path.Path) *np.Rerror {
	name := path.Base()
	if name == "." {
		return np.MkRerror(fcall.MkErr(fcall.TErrInval, name))
	}

	// lock path to make WatchV and Remove interact correctly
	dlk := ps.plt.Acquire(ctx, path.Dir())
	flk := ps.plt.Acquire(ctx, path)
	defer ps.plt.ReleaseLocks(ctx, dlk, flk)

	ps.stats.IncPathString(flk.Path())

	db.DPrintf("PROTSRV", "%v: removeObj %v in %v", ctx.Uname(), name, o)

	// Call before Remove(), because after remove o's underlying
	// object may not exist anymore.
	ephemeral := o.Perm().IsEphemeral()
	err := o.Parent().Remove(ctx, name)
	if err != nil {
		return np.MkRerror(err)
	}

	ps.vt.IncVersion(o.Path())
	ps.vt.IncVersion(o.Parent().Path())

	ps.wt.WakeupWatch(flk)
	ps.wt.WakeupWatch(dlk)

	if ephemeral {
		ps.et.Del(o)
	}
	return nil
}

// Remove for backwards compatability; SigmaOS uses RemoveFile (see
// below) instead of Remove, but proxy will use it.
func (ps *ProtSrv) Remove(args *np.Tremove, rets *np.Rremove) *np.Rerror {
	f, err := ps.ft.Lookup(args.Fid)
	if err != nil {
		return np.MkRerror(err)
	}
	db.DPrintf("PROTSRV", "%v: Remove %v", f.Pobj().Ctx().Uname(), f.Pobj().Path())
	return ps.removeObj(f.Pobj().Ctx(), f.Pobj().Obj(), f.Pobj().Path())
}

func (ps *ProtSrv) Stat(args *np.Tstat, rets *np.Rstat) *np.Rerror {
	f, err := ps.ft.Lookup(args.Tfid())
	if err != nil {
		return np.MkRerror(err)
	}
	db.DPrintf("PROTSRV", "%v: Stat %v", f.Pobj().Ctx().Uname(), f)
	o := f.Pobj().Obj()
	st, r := o.Stat(f.Pobj().Ctx())
	if r != nil {
		return np.MkRerror(r)
	}
	rets.Stat = st
	return nil
}

//
// Rename: within the same directory (Wstat) and rename across directories
//

func (ps *ProtSrv) Wstat(args *np.Twstat, rets *np.Rwstat) *np.Rerror {
	f, err := ps.ft.Lookup(args.Fid)
	if err != nil {
		return np.MkRerror(err)
	}
	db.DPrintf("PROTSRV", "%v: Wstat %v %v", f.Pobj().Ctx().Uname(), f, args)
	o := f.Pobj().Obj()
	if args.Stat.Name != "" {
		// update Name atomically with rename

		dst := f.Pobj().Path().Dir().Copy().AppendPath(path.Split(args.Stat.Name))

		dlk, slk := ps.plt.AcquireLocks(f.Pobj().Ctx(), f.Pobj().Path().Dir(), f.Pobj().Path().Base())
		defer ps.plt.ReleaseLocks(f.Pobj().Ctx(), dlk, slk)
		tlk := ps.plt.Acquire(f.Pobj().Ctx(), dst)
		defer ps.plt.Release(f.Pobj().Ctx(), tlk)

		err := o.Parent().Rename(f.Pobj().Ctx(), f.Pobj().Path().Base(), args.Stat.Name)
		if err != nil {
			return np.MkRerror(err)
		}
		ps.vt.IncVersion(f.Pobj().Obj().Path())
		ps.wt.WakeupWatch(tlk) // trigger create watch
		ps.wt.WakeupWatch(slk) // trigger remove watch
		ps.wt.WakeupWatch(dlk) // trigger dir watch
		f.Pobj().SetPath(dst)
	}
	// XXX ignore other Wstat for now
	return nil
}

// d1 first?
func lockOrder(d1 fs.FsObj, d2 fs.FsObj) bool {
	if d1.Path() < d2.Path() {
		return true
	} else if d1.Path() == d2.Path() { // would have used wstat instead of renameat
		db.DFatalf("lockOrder")
		return false
	} else {
		return false
	}
}

func (ps *ProtSrv) Renameat(args *np.Trenameat, rets *np.Rrenameat) *np.Rerror {
	oldf, err := ps.ft.Lookup(args.OldFid)
	if err != nil {
		return np.MkRerror(err)
	}
	newf, err := ps.ft.Lookup(args.NewFid)
	if err != nil {
		return np.MkRerror(err)
	}
	db.DPrintf("PROTSRV", "%v: renameat %v %v %v", oldf.Pobj().Ctx().Uname(), oldf, newf, args)
	oo := oldf.Pobj().Obj()
	no := newf.Pobj().Obj()
	switch d1 := oo.(type) {
	case fs.Dir:
		d2, ok := no.(fs.Dir)
		if !ok {
			return np.MkRerror(fcall.MkErr(fcall.TErrNotDir, newf.Pobj().Path()))
		}
		if oo.Path() == no.Path() {
			return np.MkRerror(fcall.MkErr(fcall.TErrInval, newf.Pobj().Path()))
		}

		var d1lk, d2lk, srclk, dstlk *lockmap.PathLock
		if srcfirst := lockOrder(oo, no); srcfirst {
			d1lk, srclk = ps.plt.AcquireLocks(oldf.Pobj().Ctx(), oldf.Pobj().Path(), args.OldName)
			d2lk, dstlk = ps.plt.AcquireLocks(newf.Pobj().Ctx(), newf.Pobj().Path(), args.NewName)
		} else {
			d2lk, dstlk = ps.plt.AcquireLocks(newf.Pobj().Ctx(), newf.Pobj().Path(), args.NewName)
			d1lk, srclk = ps.plt.AcquireLocks(oldf.Pobj().Ctx(), oldf.Pobj().Path(), args.OldName)
		}
		defer ps.plt.ReleaseLocks(oldf.Pobj().Ctx(), d1lk, srclk)
		defer ps.plt.ReleaseLocks(newf.Pobj().Ctx(), d2lk, dstlk)

		err := d1.Renameat(oldf.Pobj().Ctx(), args.OldName, d2, args.NewName)
		if err != nil {
			return np.MkRerror(err)
		}
		ps.vt.IncVersion(newf.Pobj().Obj().Path())
		ps.vt.IncVersion(oldf.Pobj().Obj().Path())

		ps.wt.WakeupWatch(dstlk) // trigger create watch
		ps.wt.WakeupWatch(srclk) // trigger remove watch
		ps.wt.WakeupWatch(d1lk)  // trigger one dir watch
		ps.wt.WakeupWatch(d2lk)  // trigger the other dir watch
	default:
		return np.MkRerror(fcall.MkErr(fcall.TErrNotDir, oldf.Pobj().Path()))
	}
	return nil
}

func (ps *ProtSrv) lookupWalk(fid np.Tfid, wnames path.Path, resolve bool) (*fid.Fid, path.Path, fs.FsObj, *fcall.Err) {
	f, err := ps.ft.Lookup(fid)
	if err != nil {
		return nil, nil, nil, err
	}
	lo := f.Pobj().Obj()
	fname := append(f.Pobj().Path(), wnames...)
	if len(wnames) > 0 {
		lo, err = ps.lookupObjLast(f.Pobj().Ctx(), f, wnames, resolve)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	return f, fname, lo, nil
}

func (ps *ProtSrv) lookupWalkOpen(fid np.Tfid, wnames path.Path, resolve bool, mode np.Tmode) (*fid.Fid, path.Path, fs.FsObj, fs.File, *fcall.Err) {
	f, fname, lo, err := ps.lookupWalk(fid, wnames, resolve)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	ps.stats.IncPath(fname)
	no, err := lo.Open(f.Pobj().Ctx(), mode)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if no != nil {
		lo = no
	}
	i, err := fs.Obj2File(lo, fname)
	if err != nil {
		lo.Close(f.Pobj().Ctx(), mode)
		return nil, nil, nil, nil, err
	}
	return f, fname, lo, i, nil
}

func (ps *ProtSrv) RemoveFile(args *np.Tremovefile, rets *np.Rremove) *np.Rerror {
	f, fname, lo, err := ps.lookupWalk(args.Fid, args.Wnames, args.Resolve)
	if err != nil {
		return np.MkRerror(err)
	}
	db.DPrintf("PROTSRV", "%v: RemoveFile %v %v %v", f.Pobj().Ctx().Uname(), f.Pobj().Path(), fname, args.Fid)
	return ps.removeObj(f.Pobj().Ctx(), lo, fname)
}

func (ps *ProtSrv) GetFile(args *np.Tgetfile, rets *np.Rread) *np.Rerror {
	if args.Tcount() > np.MAXGETSET {
		return np.MkRerror(fcall.MkErr(fcall.TErrInval, "too large"))
	}
	f, fname, lo, i, err := ps.lookupWalkOpen(args.Tfid(), args.Wnames, args.Resolve, args.Tmode())
	if err != nil {
		return np.MkRerror(err)
	}
	db.DPrintf("PROTSRV", "GetFile f %v args %v %v", f.Pobj().Ctx().Uname(), args, fname)
	rets.Data, err = i.Read(f.Pobj().Ctx(), args.Toffset(), args.Tcount(), np.NoV)
	if err != nil {
		return np.MkRerror(err)
	}
	if err := lo.Close(f.Pobj().Ctx(), args.Tmode()); err != nil {
		return np.MkRerror(err)
	}
	return nil
}

func (ps *ProtSrv) SetFile(args *np.Tsetfile, rets *np.Rwrite) *np.Rerror {
	if np.Tsize(len(args.Data)) > np.MAXGETSET {
		return np.MkRerror(fcall.MkErr(fcall.TErrInval, "too large"))
	}
	f, fname, lo, i, err := ps.lookupWalkOpen(args.Fid, args.Wnames, args.Resolve, args.Mode)
	if err != nil {
		return np.MkRerror(err)
	}

	db.DPrintf("PROTSRV", "SetFile f %v args %v %v", f.Pobj().Ctx().Uname(), args, fname)

	if args.Mode&np.OAPPEND == np.OAPPEND && args.Offset != np.NoOffset {
		return np.MkRerror(fcall.MkErr(fcall.TErrInval, "offset should be np.NoOffset"))
	}
	if args.Offset == np.NoOffset && args.Mode&np.OAPPEND != np.OAPPEND {
		return np.MkRerror(fcall.MkErr(fcall.TErrInval, "mode shouldbe OAPPEND"))
	}

	n, err := i.Write(f.Pobj().Ctx(), args.Offset, args.Data, np.NoV)
	if err != nil {
		return np.MkRerror(err)
	}

	if err := lo.Close(f.Pobj().Ctx(), args.Mode); err != nil {
		return np.MkRerror(err)
	}
	rets.Count = uint32(n)
	return nil
}

func (ps *ProtSrv) PutFile(args *np.Tputfile, rets *np.Rwrite) *np.Rerror {
	if np.Tsize(len(args.Data)) > np.MAXGETSET {
		return np.MkRerror(fcall.MkErr(fcall.TErrInval, "too large"))
	}
	// walk to directory
	f, dname, lo, err := ps.lookupWalk(args.Fid, args.Wnames[0:len(args.Wnames)-1], false)
	if err != nil {
		return np.MkRerror(err)
	}
	fn := append(f.Pobj().Path(), args.Wnames...)

	db.DPrintf("PROTSRV", "%v: PutFile o %v args %v (%v)", f.Pobj().Ctx().Uname(), f, args, dname)

	if !lo.Perm().IsDir() {
		return np.MkRerror(fcall.MkErr(fcall.TErrNotDir, dname))
	}
	dlk := ps.plt.Acquire(f.Pobj().Ctx(), dname)
	defer ps.plt.Release(f.Pobj().Ctx(), dlk)

	// flk also ensures that two Puts execute atomically
	lo, flk, err := ps.createObj(f.Pobj().Ctx(), lo.(fs.Dir), dlk, fn, args.Perm, args.Mode)
	if err != nil {
		return np.MkRerror(err)
	}
	defer ps.plt.Release(f.Pobj().Ctx(), flk)
	qid := ps.mkQid(lo.Perm(), lo.Path())
	f = ps.makeFid(f.Pobj().Ctx(), dname, fn.Base(), lo, args.Perm.IsEphemeral(), qid)
	i, err := fs.Obj2File(lo, fn)
	if err != nil {
		return np.MkRerror(err)
	}
	n, err := i.Write(f.Pobj().Ctx(), args.Offset, args.Data, np.NoV)
	if err != nil {
		return np.MkRerror(err)
	}
	err = lo.Close(f.Pobj().Ctx(), args.Mode)
	if err != nil {
		return np.MkRerror(err)
	}
	rets.Count = uint32(n)
	return nil
}

func (ps *ProtSrv) Snapshot() []byte {
	return ps.snapshot()
}
