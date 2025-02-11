package fslib

import (
	sp "sigmaos/sigmap"
)

func (fl *FsLib) MakePipe(name string, lperm sp.Tperm) error {
	lperm = lperm | sp.DMNAMEDPIPE
	// ORDWR so that close doesn't do anything to the pipe state
	fd, err := fl.Create(name, lperm, sp.ORDWR)
	if err != nil {
		return err
	}
	return fl.Close(fd)
}
