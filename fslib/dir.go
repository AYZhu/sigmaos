package fslib

import (
	"fmt"
	"path"

	db "sigmaos/debug"
	"sigmaos/reader"
	"sigmaos/serr"
	sp "sigmaos/sigmap"
)

func (fl *FsLib) MkDir(path string, perm sp.Tperm) error {
	perm = perm | sp.DMDIR
	fd, err := fl.Create(path, perm, sp.OREAD)
	if err != nil {
		return err
	}
	fl.Close(fd)
	return nil
}

func (fl *FsLib) IsDir(name string) (bool, error) {
	st, err := fl.FdClient.Stat(name)
	if err != nil {
		return false, err
	}
	return st.Tmode().IsDir(), nil
}

// Too stop early, f must return true.  Returns true if stopped early.
func (fl *FsLib) ProcessDir(dir string, f func(*sp.Stat) (bool, error)) (bool, error) {
	rdr, err := fl.OpenReader(dir)
	if err != nil {
		return false, err
	}
	defer rdr.Close()
	return reader.ReadDir(reader.NewDirReader(rdr), f)
}

func (fl *FsLib) GetDir(dir string) ([]*sp.Stat, error) {
	st, rdr, err := fl.ReadDir(dir)
	if rdr != nil {
		rdr.Close()
	}
	return st, err
}

// Also returns reader.Reader for ReadDirWatch
func (fl *FsLib) ReadDir(dir string) ([]*sp.Stat, *reader.Reader, error) {
	rdr, err := fl.OpenReader(dir)
	if err != nil {
		return nil, nil, err
	}
	dirents := []*sp.Stat{}
	_, error := reader.ReadDir(reader.NewDirReader(rdr), func(st *sp.Stat) (bool, error) {
		dirents = append(dirents, st)
		return false, nil
	})
	return dirents, rdr, error
}

// XXX should use Reader
func (fl *FsLib) CopyDir(src, dst string) error {
	_, err := fl.ProcessDir(src, func(st *sp.Stat) (bool, error) {
		s := src + "/" + st.Name
		d := dst + "/" + st.Name
		db.DPrintf(db.FSLIB, "CopyFile: %v %v\n", s, d)
		b, err := fl.GetFile(s)
		if err != nil {
			return true, err
		}
		_, err = fl.PutFile(d, 0777, sp.OWRITE, b)
		if err != nil {
			return true, err
		}
		return false, nil
	})
	return err
}

func (fl *FsLib) MoveFiles(src, dst string) (int, error) {
	sts, err := fl.GetDir(src) // XXX handle one entry at the time?
	if err != nil {
		return 0, err
	}
	n := 0
	for _, st := range sts {
		db.DPrintf(db.ALWAYS, "move %v to %v\n", st.Name, dst)
		to := dst + "/" + st.Name
		if fl.Rename(src+"/"+st.Name, to) != nil {
			return n, err
		}
		n += 1
	}
	return n, nil
}

func (fsl *FsLib) RmDir(dir string) error {
	if err := fsl.RmDirEntries(dir); err != nil {
		return err
	}
	return fsl.Remove(dir)
}

func (fsl *FsLib) RmDirEntries(dir string) error {
	sts, err := fsl.GetDir(dir)
	if err != nil {
		return err
	}
	for _, st := range sts {
		if st.Tmode().IsDir() {
			if err := fsl.RmDir(dir + "/" + st.Name); err != nil {
				return err
			}
		} else {
			if err := fsl.Remove(dir + "/" + st.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (fsl *FsLib) SprintfDir(d string) (string, error) {
	return fsl.sprintfDirIndent(d, "")
}

func (fsl *FsLib) sprintfDirIndent(d string, indent string) (string, error) {
	s := fmt.Sprintf("%v dir %v\n", indent, d)
	sts, err := fsl.GetDir(d)
	if err != nil {
		return "", err
	}
	for _, st := range sts {
		s += fmt.Sprintf("%v %v %v\n", indent, st.Name, st.Qid.Type)
		if st.Tmode().IsDir() {
			s1, err := fsl.sprintfDirIndent(d+"/"+st.Name, indent+" ")
			if err != nil {
				return s, err
			}
			s += s1
		}
	}
	return s, nil
}

func Present(sts []*sp.Stat, names []string) bool {
	n := 0
	m := make(map[string]bool)
	for _, n := range names {
		m[path.Base(n)] = true
	}
	for _, st := range sts {
		if _, ok := m[st.Name]; ok {
			n += 1
		}
	}
	return n == len(names)
}

type Fwait func([]*sp.Stat) bool

// Keep reading dir until wait returns false (e.g., a new file has
// been created in dir)
func (fsl *FsLib) ReadDirWatch(dir string, wait Fwait) ([]*sp.Stat, error) {
	for {
		sts, rdr, err := fsl.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		if wait(sts) { // wait for new inputs?
			ch := make(chan error)
			if err := fsl.SetDirWatch(rdr.Fid(), dir, func(p string, r error) {
				ch <- r
			}); err != nil {
				rdr.Close()
				if serr.IsErrCode(err, serr.TErrVersion) {
					db.DPrintf(db.ALWAYS, "ReadDirWatch: Version mismatch %v\n", dir)
					continue
				}
				return nil, err
			}
			if err := <-ch; err != nil {
				rdr.Close()
				if serr.IsErrCode(err, serr.TErrVersion) {
					db.DPrintf(db.ALWAYS, "ReadDirWatch: Version mismatch %v\n", dir)
					continue
				}
				return nil, err
			}
			rdr.Close()
			continue // read again
		}
		rdr.Close()
		return sts, nil
	}
	return nil, nil
}
