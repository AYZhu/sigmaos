package reader

import (
	"io"

	db "sigmaos/debug"
	"sigmaos/fcall"
	"sigmaos/fidclnt"
	np "sigmaos/sigmap"
)

type Reader struct {
	fc     *fidclnt.FidClnt
	path   string
	fid    np.Tfid
	off    np.Toffset
	eof    bool
	fenced bool
}

func (rdr *Reader) Path() string {
	return rdr.path
}

func (rdr *Reader) Fid() np.Tfid {
	return rdr.fid
}

func (rdr *Reader) Nbytes() np.Tlength {
	return np.Tlength(rdr.off)
}

func (rdr *Reader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if rdr.eof {
		return 0, io.EOF
	}
	var b []byte
	var err *fcall.Err
	sz := np.Tsize(len(p))
	if rdr.fenced {
		b, err = rdr.fc.ReadV(rdr.fid, rdr.off, sz, np.NoV)
	} else {
		b, err = rdr.fc.ReadVU(rdr.fid, rdr.off, sz, np.NoV)
	}
	if err != nil {
		db.DPrintf("READER_ERR", "Read %v err %v\n", rdr.path, err)
		return 0, err
	}
	if len(b) == 0 {
		rdr.eof = true
		return 0, io.EOF
	}
	if len(p) != len(b) {
		db.DPrintf("READER_ERR", "Read short %v %v %v\n", rdr.path, len(p), len(b))
	}
	// XXX change rdr.Read to avoid copy
	copy(p, b)
	rdr.off += np.Toffset(len(b))
	return len(b), nil
}

func (rdr *Reader) GetData() ([]byte, error) {
	b, err := rdr.GetDataErr()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (rdr *Reader) GetDataErr() ([]byte, *fcall.Err) {
	return rdr.fc.ReadV(rdr.fid, 0, np.MAXGETSET, np.NoV)
}

func (rdr *Reader) Lseek(o np.Toffset) error {
	rdr.off = o
	return nil
}

func (rdr *Reader) Close() error {
	err := rdr.fc.Clunk(rdr.fid)
	if err != nil {
		return err
	}
	return nil
}

func (rdr *Reader) Unfence() {
	rdr.fenced = false
}

func MakeReader(fc *fidclnt.FidClnt, path string, fid np.Tfid, chunksz np.Tsize) *Reader {
	return &Reader{fc, path, fid, 0, false, true}
}
