package fslib

import (
	sp "sigmaos/sigmap"
)

func (fl *FsLib) Symlink(target []byte, link string, lperm sp.Tperm) error {
	lperm = lperm | sp.DMSYMLINK
	fd, err := fl.Create(link, lperm, sp.OWRITE)
	if err != nil {
		return err
	}
	_, err = fl.Write(fd, target)
	if err != nil {
		return err
	}
	return fl.Close(fd)
}
