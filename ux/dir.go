package fsux

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"

	db "sigmaos/debug"
	"sigmaos/fs"
	"sigmaos/path"
	"sigmaos/serr"
	sp "sigmaos/sigmap"
	"sigmaos/sorteddir"
)

type Dir struct {
	*Obj
	sd *sorteddir.SortedDir
}

func (d *Dir) String() string {
	return fmt.Sprintf("o %v sd %v", d.Obj, d.sd)
}

func makeDir(path path.Path) (*Dir, *serr.Err) {
	d := &Dir{}
	o, err := makeObj(path)
	if err != nil {
		return nil, err
	}
	d.Obj = o
	d.sd = sorteddir.MkSortedDir()
	return d, nil
}

func (d *Dir) uxReadDir() *serr.Err {
	dirents, err := ioutil.ReadDir(d.PathName())
	if err != nil {
		return serr.UxErrnoToErr(err, d.pathName.Base())
	}
	for _, e := range dirents {
		if st, err := ustat(d.pathName.Copy().Append(e.Name())); err != nil {
			// another proc may have removed the file
			continue
		} else {
			d.sd.Insert(st.Name, st)
		}
	}
	db.DPrintf(db.UX, "%v: uxReadDir %v\n", d, d.sd.Len())
	return nil
}

func (d *Dir) ReadDir(ctx fs.CtxI, cursor int, cnt sp.Tsize) ([]*sp.Stat, *serr.Err) {
	db.DPrintf(db.UX, "%v: ReadDir %v %v %v\n", ctx, d, cursor, cnt)
	dents := make([]*sp.Stat, 0, d.sd.Len())
	d.sd.Iter(func(n string, e interface{}) bool {
		dents = append(dents, e.(*sp.Stat))
		return true
	})
	if cursor > len(dents) {
		return nil, nil
	} else {
		return dents[cursor:], nil
	}
}

func (d *Dir) Open(ctx fs.CtxI, m sp.Tmode) (fs.FsObj, *serr.Err) {
	if err := d.uxReadDir(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (d *Dir) Close(ctx fs.CtxI, mode sp.Tmode) *serr.Err {
	d.sd = sorteddir.MkSortedDir()
	return nil
}

// XXX O_CREATE/O_EXCL
func (d *Dir) mkDir(ctx fs.CtxI, name string, perm sp.Tperm, m sp.Tmode) (*Dir, *serr.Err) {
	p := d.pathName.Append(name).String()
	error := os.Mkdir(p, os.FileMode(perm&0777))
	if error != nil {
		return nil, serr.UxErrnoToErr(error, name)
	}
	d1, err := makeDir(append(d.pathName, name))
	if err != nil {
		return nil, err
	}
	return d1, nil
}

func (d *Dir) mkFile(ctx fs.CtxI, name string, perm sp.Tperm, m sp.Tmode) (fs.FsObj, *serr.Err) {
	p := d.pathName.Append(name).String()
	fd, error := syscall.Open(p, uxFlags(m)|syscall.O_CREAT|syscall.O_EXCL, uint32(perm&0777))
	if error != nil {
		return nil, serr.UxErrnoToErr(error, name)
	}
	f, err := makeFile(append(d.pathName, name))
	if err != nil {
		return nil, err
	}
	f.fd = fd
	return f, nil
}

func (d *Dir) mkPipe(ctx fs.CtxI, name string, perm sp.Tperm, m sp.Tmode) (fs.FsObj, *serr.Err) {
	p := d.pathName.Append(name).String()
	error := syscall.Mkfifo(p, uint32(perm&0777))
	if error != nil {
		return nil, serr.UxErrnoToErr(error, name)
	}
	f, err := makePipe(ctx, append(d.pathName, name))
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (d *Dir) mkSym(ctx fs.CtxI, name string, perm sp.Tperm, m sp.Tmode) (fs.FsObj, *serr.Err) {
	p := d.pathName.Append(name)
	log.Printf("mkSym %s\n", p)
	s, err := makeSymlink(p, true)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// XXX how to delete ephemeral files after crash
func (d *Dir) Create(ctx fs.CtxI, name string, perm sp.Tperm, m sp.Tmode, lid sp.TleaseId, f sp.Tfence) (fs.FsObj, *serr.Err) {
	db.DPrintf(db.UX, "%v: Create %v n %v perm %v m %v\n", ctx, d, name, perm, m)
	if perm.IsDir() {
		return d.mkDir(ctx, name, perm, m)
	} else if perm.IsPipe() {
		return d.mkPipe(ctx, name, perm, m)
	} else if perm.IsSymlink() {
		return d.mkSym(ctx, name, perm, m)
	} else {
		return d.mkFile(ctx, name, perm, m)
	}
}

func (d *Dir) LookupPath(ctx fs.CtxI, path path.Path) ([]fs.FsObj, fs.FsObj, path.Path, *serr.Err) {
	name := path[0]
	db.DPrintf(db.UX, "%v: Lookup %v %v\n", ctx, d, name)
	st, err := ustat(d.pathName.Append(name))
	if err != nil {
		db.DPrintf(db.UX, "%v: Lookup %v %v err %v\n", ctx, d, name, err)
		return nil, nil, path, err
	}
	db.DPrintf(db.UX, "%v: Lookup %v %v st %v\n", ctx, d, name, st)
	var o fs.FsObj
	if st.Tmode().IsDir() {
		o, err = makeDir(append(d.pathName, name))
		if err != nil {
			return nil, nil, path, err
		}
	} else if st.Tmode().IsSymlink() {
		o, err = makeSymlink(append(d.pathName, name), false)
		if err != nil {
			return nil, nil, path, err
		}
	} else if st.Tmode().IsPipe() {
		o, err = makePipe(ctx, append(d.pathName, name))
		if err != nil {
			return nil, nil, path, err
		}
	} else {
		o, err = makeFile(append(d.pathName, name))
		if err != nil {
			return nil, nil, path, err
		}
	}
	return []fs.FsObj{o}, o, path[1:], nil
}

func (d *Dir) Renameat(ctx fs.CtxI, from string, dd fs.Dir, to string, f sp.Tfence) *serr.Err {
	oldPath := d.PathName() + "/" + from
	newPath := dd.(*Dir).PathName() + "/" + to
	db.DPrintf(db.UX, "%v: Renameat d:%v from:%v to:%v\n", ctx, d, from, to)
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return serr.UxErrnoToErr(err, to)
	}
	return nil
}

func (d *Dir) Remove(ctx fs.CtxI, name string, f sp.Tfence) *serr.Err {
	db.DPrintf(db.UX, "%v: Remove %v %v\n", ctx, d, name)
	p := d.pathName.Copy().Append(name)
	o, err := makeObj(p)
	if err != nil {
		return err
	}
	error := os.Remove(p.String())
	if error != nil {
		return serr.UxErrnoToErr(error, name)
	}
	if o.Perm().IsPipe() {
		pipe := fsux.ot.AllocRef(o.path, nil)
		if pipe != nil {
			pipe.(*Pipe).Unlink()
		}
	}
	return nil
}

func (d *Dir) Rename(ctx fs.CtxI, from, to string, f sp.Tfence) *serr.Err {
	oldPath := d.PathName() + "/" + from
	newPath := d.PathName() + "/" + to
	db.DPrintf(db.UX, "%v: Rename d:%v from:%v to:%v\n", ctx, d, from, to)
	error := os.Rename(oldPath, newPath)
	if error != nil {
		return serr.UxErrnoToErr(error, to)
	}
	// XXX unlink on pipe, if pipe
	return nil
}
