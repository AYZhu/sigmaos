package named

import (
	db "sigmaos/debug"
	"sigmaos/fs"
	"sigmaos/serr"
	"sigmaos/sessp"
	sp "sigmaos/sigmap"
)

type File struct {
	*Obj
}

func makeFile(o *Obj) *File {
	f := &File{}
	f.Obj = o
	return f
}

func (f *File) Open(ctx fs.CtxI, m sp.Tmode) (fs.FsObj, *serr.Err) {
	db.DPrintf(db.NAMED, "%v: FileOpen %v m 0x%x path %v\n", ctx, f, m, f.Obj.pn)
	return nil, nil
}

func (f *File) Close(ctx fs.CtxI, mode sp.Tmode) *serr.Err {
	db.DPrintf(db.NAMED, "%v: FileClose %v\n", ctx, f)
	return nil
}

// XXX maybe do get
func (f *File) Read(ctx fs.CtxI, offset sp.Toffset, n sessp.Tsize, v sp.TQversion) ([]byte, *serr.Err) {
	db.DPrintf(db.NAMED, "%v: Read: %v off %v cnt %v\n", ctx, f, offset, n)
	if offset >= f.LenOff() {
		return nil, nil
	} else {
		// XXX overflow?
		end := offset + sp.Toffset(n)
		if end >= f.LenOff() {
			end = f.LenOff()
		}
		b := f.data[offset:end]
		return b, nil
	}
}

func (f *File) LenOff() sp.Toffset {
	return sp.Toffset(len(f.Obj.data))
}

func (f *File) Write(ctx fs.CtxI, offset sp.Toffset, b []byte, v sp.TQversion) (sessp.Tsize, *serr.Err) {
	db.DPrintf(db.NAMED, "%v: Write: off %v cnt %v\n", f, offset, len(b))
	cnt := sessp.Tsize(len(b))
	sz := sp.Toffset(len(b))

	if offset == sp.NoOffset { // OAPPEND
		offset = f.LenOff()
	}

	if offset >= f.LenOff() { // passed end of file?
		n := f.LenOff() - offset

		f.Obj.data = append(f.Obj.data, make([]byte, n)...)
		f.Obj.data = append(f.Obj.data, b...)

		if err := f.Obj.putObj(); err != nil {
			return 0, err
		}

		return cnt, nil
	}

	var d []byte
	if offset+sz < f.LenOff() { // in the middle of the file?
		d = f.Obj.data[offset+sz:]
	}
	f.Obj.data = f.Obj.data[0:offset]
	f.Obj.data = append(f.Obj.data, b...)
	f.Obj.data = append(f.data, d...)

	if err := f.Obj.putObj(); err != nil {
		return 0, err
	}
	return sessp.Tsize(len(b)), nil
}
