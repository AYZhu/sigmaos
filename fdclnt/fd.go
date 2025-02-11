package fdclnt

import (
	"sync"

	"sigmaos/serr"
	sp "sigmaos/sigmap"
)

const (
	MAXFD = 20
)

type FdState struct {
	offset sp.Toffset
	fid    sp.Tfid
	mode   sp.Tmode
}

type FdTable struct {
	sync.Mutex
	fds     []FdState
	freefds map[int]bool
}

func mkFdTable() *FdTable {
	fdt := &FdTable{}
	fdt.fds = make([]FdState, 0, MAXFD)
	fdt.freefds = make(map[int]bool)
	return fdt
}

func (fdt *FdTable) allocFd(nfid sp.Tfid, m sp.Tmode) int {
	fdt.Lock()
	defer fdt.Unlock()

	if len(fdt.freefds) > 0 {
		for i, _ := range fdt.freefds {
			delete(fdt.freefds, i)
			fdt.fds[i].offset = 0
			fdt.fds[i].fid = nfid
			fdt.fds[i].mode = m
			return i
		}
	}

	// no free one
	fdt.fds = append(fdt.fds, FdState{0, nfid, m})
	return len(fdt.fds) - 1
}

func (fdt *FdTable) closefd(fd int) {
	fdt.Lock()
	defer fdt.Unlock()

	fdt.fds[fd].fid = sp.NoFid
	fdt.freefds[fd] = true
}

// Caller must have locked fdt
func (fdt *FdTable) lookupL(fd int) (*FdState, *serr.Err) {
	if fd < 0 || fd >= len(fdt.fds) {
		return nil, serr.MkErr(serr.TErrBadFd, fd)
	}
	if fdt.fds[fd].fid == sp.NoFid {
		return nil, serr.MkErr(serr.TErrBadFd, fd)
	}
	return &fdt.fds[fd], nil
}

func (fdt *FdTable) lookup(fd int) (sp.Tfid, *serr.Err) {
	fdt.Lock()
	defer fdt.Unlock()

	st, err := fdt.lookupL(fd)
	if err != nil {
		return sp.NoFid, err
	}
	return st.fid, nil
}

func (fdt *FdTable) lookupOff(fd int) (sp.Tfid, sp.Toffset, *serr.Err) {
	fdt.Lock()
	defer fdt.Unlock()

	st, err := fdt.lookupL(fd)
	if err != nil {
		return sp.NoFid, 0, err
	}
	return st.fid, st.offset, nil
}

func (fdt *FdTable) setOffset(fd int, off sp.Toffset) *serr.Err {
	fdt.Lock()
	defer fdt.Unlock()

	st, err := fdt.lookupL(fd)
	if err != nil {
		return err
	}
	st.offset = off
	return nil
}

func (fdt *FdTable) incOff(fd int, off sp.Toffset) *serr.Err {
	fdt.Lock()
	defer fdt.Unlock()

	st, err := fdt.lookupL(fd)
	if err != nil {
		return err
	}
	st.offset += off
	return nil
}
