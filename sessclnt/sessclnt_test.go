package sessclnt_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"sigmaos/config"
	db "sigmaos/debug"
	"sigmaos/fslib"
	"sigmaos/groupmgr"
	"sigmaos/kvgrp"
	"sigmaos/proc"
	"sigmaos/semclnt"
	"sigmaos/serr"
	sp "sigmaos/sigmap"
	"sigmaos/test"
)

const (
	CRASH     = 1000
	PARTITION = 200
	NETFAIL   = 200
	NTRIALS   = "3001"
	JOBDIR    = "name/group"
	GRP       = "grp-0"
)

type Tstate struct {
	*test.Tstate
	grp string
	gm  *groupmgr.GroupMgr
}

func makeTstate(t *testing.T, ncrash, crash, partition, netfail int) *Tstate {
	ts := &Tstate{grp: GRP}
	ts.Tstate = test.MakeTstateAll(t)
	ts.RmDir(JOBDIR)
	ts.MkDir(JOBDIR, 0777)
	ts.gm = groupmgr.Start(ts.SigmaClnt, 0, "kvd", []string{ts.grp, strconv.FormatBool(test.Overlays)}, JOBDIR, 0, ncrash, crash, partition, netfail)
	cfg, err := kvgrp.WaitStarted(ts.SigmaClnt.FsLib, JOBDIR, ts.grp)
	assert.Nil(t, err)
	db.DPrintf(db.TEST, "cfg %v\n", cfg)
	return ts
}

// Server crashes store a semaphore, groupmgr starts a new server,
// which will return a not-found for the semaphore, which is
// interpreted as a successful down by the semclnt.
func TestServerCrash(t *testing.T) {
	ts := makeTstate(t, 1, CRASH, 0, 0)

	sem := semclnt.MakeSemClnt(ts.FsLib, kvgrp.GrpPath(JOBDIR, ts.grp)+"/sem")
	err := sem.Init(0)
	assert.Nil(t, err)

	db.DPrintf(db.TEST, "Sem %v", kvgrp.GrpPath(JOBDIR, ts.grp)+"/sem")

	ch := make(chan error)
	go func() {
		scfg := config.NewAddedSigmaConfig(ts.SigmaConfig(), 1)
		fsl, err := fslib.MakeFsLib(scfg)
		assert.Nil(t, err)
		sem := semclnt.MakeSemClnt(fsl, kvgrp.GrpPath(JOBDIR, ts.grp)+"/sem")
		err = sem.Down()
		ch <- err
	}()

	err = <-ch
	assert.Nil(ts.T, err, "down")

	ts.gm.Stop()

	ts.Shutdown()
}

func BurstProc(n int, f func(chan error)) error {
	ch := make(chan error)
	for i := 0; i < n; i++ {
		go f(ch)
	}
	var err error
	for i := 0; i < n; i++ {
		r := <-ch
		if r != nil && err != nil {
			err = r
		}
	}
	return err
}

func TestProcManyOK(t *testing.T) {
	ts := test.MakeTstateAll(t)
	a := proc.MakeProc("proctest", []string{NTRIALS, "sleeper", "1us", ""})
	err := ts.Spawn(a)
	assert.Nil(t, err, "Spawn")
	err = ts.WaitStart(a.GetPid())
	assert.Nil(t, err, "WaitStart error")
	status, err := ts.WaitExit(a.GetPid())
	assert.Nil(t, err, "waitexit")
	assert.True(t, status.IsStatusOK(), status)
	ts.Shutdown()
}

func TestProcCrashMany(t *testing.T) {
	ts := test.MakeTstateAll(t)
	a := proc.MakeProc("proctest", []string{NTRIALS, "crash"})
	err := ts.Spawn(a)
	assert.Nil(t, err, "Spawn")
	err = ts.WaitStart(a.GetPid())
	assert.Nil(t, err, "WaitStart error")
	status, err := ts.WaitExit(a.GetPid())
	assert.Nil(t, err, "waitexit")
	assert.True(t, status.IsStatusOK(), status)
	ts.Shutdown()
}

func TestProcPartitionMany(t *testing.T) {
	ts := test.MakeTstateAll(t)
	a := proc.MakeProc("proctest", []string{NTRIALS, "partition"})
	err := ts.Spawn(a)
	assert.Nil(t, err, "Spawn")
	err = ts.WaitStart(a.GetPid())
	assert.Nil(t, err, "WaitStart error")
	status, err := ts.WaitExit(a.GetPid())
	assert.Nil(t, err, "waitexit")
	if assert.NotNil(t, status, "nil status") {
		assert.True(t, status.IsStatusOK(), status)
	}
	ts.Shutdown()
}

func TestReconnectSimple(t *testing.T) {
	const N = 1000
	ts := makeTstate(t, 0, 0, 0, NETFAIL)

	ch := make(chan error)
	go func() {
		scfg := config.NewAddedSigmaConfig(ts.SigmaConfig(), 1)
		fsl, err := fslib.MakeFsLib(scfg)
		assert.Nil(t, err)
		for i := 0; i < N; i++ {
			_, err := fsl.Stat(kvgrp.GrpPath(JOBDIR, ts.grp) + "/")
			if err != nil {
				ch <- err
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
		ch <- nil
	}()

	err := <-ch
	assert.Nil(ts.T, err, "fsl1")

	ts.gm.Stop()
	ts.Shutdown()
}

func TestServerPartitionNonBlocking(t *testing.T) {
	const N = 50

	ts := makeTstate(t, 0, 0, PARTITION, 0)

	for i := 0; i < N; i++ {
		ch := make(chan error)
		go func(i int) {
			scfg := config.NewAddedSigmaConfig(ts.SigmaConfig(), i)
			fsl, err := fslib.MakeFsLib(scfg)
			assert.Nil(t, err)
			for true {
				_, err := fsl.Stat(kvgrp.GrpPath(JOBDIR, ts.grp) + "/")
				if err != nil {
					ch <- err
					break
				}
			}
			db.DPrintf(db.TEST, "Client %v done", i)
			fsl.DetachAll()
		}(i)

		err := <-ch
		assert.NotNil(ts.T, err, "stat")
	}
	db.DPrintf(db.TEST, "Stopping group")
	ts.gm.Stop()
	ts.Shutdown()
}

func TestServerPartitionBlocking(t *testing.T) {
	const N = 50

	ts := makeTstate(t, 0, 0, PARTITION, 0)

	for i := 0; i < N; i++ {
		ch := make(chan error)
		go func(i int) {
			scfg := config.NewAddedSigmaConfig(ts.SigmaConfig(), i)
			fsl, err := fslib.MakeFsLib(scfg)
			assert.Nil(t, err)
			sem := semclnt.MakeSemClnt(fsl, kvgrp.GrpPath(JOBDIR, ts.grp)+"/sem")
			sem.Init(0)
			err = sem.Down()
			ch <- err
			fsl.DetachAll()
		}(i)

		err := <-ch
		assert.NotNil(ts.T, err, "down")
	}
	ts.gm.Stop()
	ts.Shutdown()
}

const (
	FILESZ  = 50 * sp.MBYTE
	WRITESZ = 4096
)

func writer(t *testing.T, ch chan error, scfg *config.SigmaConfig) {
	fsl, err := fslib.MakeFsLib(scfg)
	assert.Nil(t, err)
	fn := sp.UX + "~local/file-" + string(scfg.Uname)
	stop := false
	nfile := 0
	for !stop {
		select {
		case <-ch:
			stop = true
		default:
			if err := fsl.Remove(fn); serr.IsErrCode(err, serr.TErrUnreachable) {
				break
			}
			w, err := fsl.CreateAsyncWriter(fn, 0777, sp.OWRITE)
			if err != nil {
				assert.True(t, serr.IsErrCode(err, serr.TErrUnreachable))
				break
			}
			nfile += 1
			buf := test.MkBuf(WRITESZ)
			if err := test.Writer(t, w, buf, FILESZ); err != nil {
				break
			}
			if err := w.Close(); err != nil {
				assert.True(t, serr.IsErrCode(err, serr.TErrUnreachable))
				break
			}
		}
	}
	assert.True(t, nfile >= 3) // a bit arbitrary
	fsl.Remove(fn)
}

func TestWriteCrash(t *testing.T) {
	const (
		N        = 20
		NCRASH   = 5
		CRASHSRV = 1000000
	)

	ts := test.MakeTstateAll(t)
	ch := make(chan error)

	for i := 0; i < N; i++ {
		scfg := config.NewAddedSigmaConfig(ts.SigmaConfig(), i)
		go writer(ts.T, ch, scfg)
	}

	crashchan := make(chan bool)
	l := &sync.Mutex{}
	for i := 0; i < NCRASH; i++ {
		go ts.CrashServer(sp.UXREL, (i+1)*CRASHSRV, l, crashchan)
	}

	for i := 0; i < NCRASH; i++ {
		<-crashchan
	}

	for i := 0; i < N; i++ {
		ch <- nil
	}

	ts.Shutdown()
}
