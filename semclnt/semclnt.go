package semclnt

import (
	"fmt"
	"time"

	db "sigmaos/debug"
	"sigmaos/fslib"
	"sigmaos/pathclnt"
	"sigmaos/serr"
	sp "sigmaos/sigmap"
)

//
// Library for binary semaphore, implemented using a file and a watch.
//

type SemClnt struct {
	path string // Path for semaphore variable
	*fslib.FsLib
}

func MakeSemClnt(fsl *fslib.FsLib, semaphore string) *SemClnt {
	c := &SemClnt{}
	c.path = semaphore
	c.FsLib = fsl
	return c
}

// Initialize semaphore variable by creating its sigmaOS state. This should
// only ever be called once globally.
func (c *SemClnt) Init(perm sp.Tperm) error {
	db.DPrintf(db.SEMCLNT, "Semaphore init %v\n", c.path)
	_, err := c.PutFile(c.path, 0777|perm, sp.OWRITE, []byte{})
	return err
}

func (c *SemClnt) InitLease(perm sp.Tperm, lid sp.TleaseId) error {
	db.DPrintf(db.SEMCLNT, "Semaphore init %v lease %v\n", c.path, lid)
	_, err := c.PutFileEphemeral(c.path, 0777|perm, sp.OWRITE, lid, []byte{})
	return err
}

// Down semaphore. If not upped yet (i.e., if file exists), block
func (c *SemClnt) Down() error {
	signal := make(chan error)
	for i := 0; i < pathclnt.MAXRETRY; i++ {
		db.DPrintf(db.SEMCLNT, "Down %d %v\n", i, c.path)
		err := c.SetRemoveWatch(c.path, func(p string, err1 error) {
			if err1 != nil {
				db.DPrintf(db.SEMCLNT_ERR, "watch %v err %v\n", c.path, err1)
			}
			signal <- err1
		})
		// If err is because file has been removed, then no error: the
		// semaphore has been "upped".
		if serr.IsErrCode(err, serr.TErrNotfound) {
			db.DPrintf(db.SEMCLNT_ERR, "down notfound %v ok err %v\n", c.path, err)
			return nil
		}
		if serr.IsErrCode(err, serr.TErrUnreachable) {
			db.DPrintf(db.SEMCLNT, "down unreachable %v ok err %v\n", c.path, err)
			time.Sleep(pathclnt.TIMEOUT * time.Millisecond)
			continue
		}
		if err == nil {
			db.DPrintf(db.SEMCLNT, "semaphore wait %v\n", c.path)
			err = <-signal
		} else {
			db.DPrintf(db.SEMCLNT_ERR, "down %v err %v\n", c.path, err)
			return err
		}
		if serr.IsErrCode(err, serr.TErrVersion) {
			db.DPrintf(db.SEMCLNT_ERR, "down %v retry err %v\n", c.path, err)
			continue
		}
		if serr.IsErrCode(err, serr.TErrUnreachable) {
			db.DPrintf(db.SEMCLNT, "down unreachable %v ok err %v\n", c.path, err)
			time.Sleep(pathclnt.TIMEOUT * time.Millisecond)
			continue
		}
		if err != nil {
			db.DPrintf(db.SEMCLNT_ERR, "down %v watch err %v\n", c.path, err)
			return err
		}
		return nil
	}
	db.DPrintf(db.SEMCLNT_ERR, "Down failed after %d retries\n", pathclnt.MAXRETRY)
	return serr.MkErr(serr.TErrUnreachable, c.path)
}

// Up a semaphore variable (i.e., remove semaphore to indicate up has
// happened).
func (c *SemClnt) Up() error {
	db.DPrintf(db.SEMCLNT, "Up %v\n", c.path)
	return c.Remove(c.path)
}

func (c *SemClnt) GetPath() string {
	return c.path
}

func (c *SemClnt) String() string {
	return fmt.Sprintf("sem %v", c.path)
}
