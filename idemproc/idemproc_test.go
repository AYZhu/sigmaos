package idemproc_test

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	db "ulambda/debug"
	"ulambda/fslib"
	"ulambda/idemproc"
	"ulambda/kernel"
	"ulambda/proc"
	"ulambda/procinit"
)

type Tstate struct {
	proc.ProcCtl
	*fslib.FsLib
	t *testing.T
	s *kernel.System
}

func makeTstate(t *testing.T) *Tstate {
	ts := &Tstate{}

	bin := ".."
	s, err := kernel.Boot(bin)
	if err != nil {
		t.Fatalf("Boot %v\n", err)
	}
	ts.s = s
	db.Name("sched_test")

	ts.FsLib = fslib.MakeFsLib("sched_test")
	ts.ProcCtl = procinit.MakeProcCtl(ts.FsLib, map[string]bool{procinit.BASESCHED: true, procinit.IDEMSCHED: true})
	ts.t = t
	return ts
}

func makeTstateNoBoot(t *testing.T, s *kernel.System) *Tstate {
	ts := &Tstate{}
	ts.t = t
	ts.s = s
	db.Name("sched_test")
	ts.FsLib = fslib.MakeFsLib("sched_test")
	ts.ProcCtl = procinit.MakeProcCtl(ts.FsLib, map[string]bool{procinit.BASESCHED: true, procinit.IDEMSCHED: true})
	return ts
}

func spawnMonitor(t *testing.T, ts *Tstate, pid string) {
	p := &idemproc.IdemProc{}
	p.Proc = &proc.Proc{pid, "bin/user/idemproc-monitor", "",
		[]string{},
		[]string{procinit.MakeProcLayers(map[string]bool{procinit.BASESCHED: true, procinit.IDEMSCHED: true})},
		proc.T_DEF, proc.C_DEF,
	}
	err := ts.Spawn(p)
	assert.Nil(t, err, "Monitor spawn")
}

func spawnSleeperlWithPid(t *testing.T, ts *Tstate, pid string) {
	p := &idemproc.IdemProc{}
	p.Proc = &proc.Proc{pid, "bin/user/sleeperl", "",
		[]string{"5s", "name/out_" + pid, ""},
		[]string{procinit.MakeProcLayers(map[string]bool{procinit.BASESCHED: true, procinit.IDEMSCHED: true})},
		proc.T_DEF, proc.C_DEF,
	}
	err := ts.Spawn(p)
	assert.Nil(t, err, "Spawn")
}

func spawnSleeperl(t *testing.T, ts *Tstate) string {
	pid := fslib.GenPid()
	spawnSleeperlWithPid(t, ts, pid)
	return pid
}

func checkSleeperlResult(t *testing.T, ts *Tstate, pid string) bool {
	res := true
	b, err := ts.ReadFile("name/out_" + pid)
	res = assert.Nil(t, err, "ReadFile") && res
	res = assert.Equal(t, string(b), "hello", "Output") && res
	return res
}

func checkSleeperlResultFalse(t *testing.T, ts *Tstate, pid string) {
	b, err := ts.ReadFile("name/out_" + pid)
	assert.NotNil(t, err, "ReadFile")
	assert.NotEqual(t, string(b), "hello", "Output")
}

func TestHelloWorld(t *testing.T) {
	ts := makeTstate(t)

	pid := spawnSleeperl(t, ts)
	time.Sleep(3 * time.Second)

	ts.s.KillOne(kernel.PROCD)

	time.Sleep(3 * time.Second)

	checkSleeperlResultFalse(t, ts, pid)

	ts.s.Shutdown(ts.FsLib)
}

func TestCrashProcd(t *testing.T) {
	ts := makeTstate(t)

	ts.s.BootProcd("..")

	N_MON := 5
	N_SLEEP := 5

	monPids := []string{}
	for i := 0; i < N_MON; i++ {
		pid := fslib.GenPid()
		spawnMonitor(t, ts, pid)
		monPids = append(monPids, pid)
	}

	time.Sleep(time.Second * 3)

	// Spawn some sleepers
	sleeperPids := []string{}
	for i := 0; i < N_SLEEP; i++ {
		pid := fslib.GenPid()
		spawnSleeperlWithPid(t, ts, pid)
		sleeperPids = append(sleeperPids, pid)
	}

	time.Sleep(time.Second * 1)

	ts.s.KillOne(kernel.PROCD)

	time.Sleep(time.Second * 10)

	for _, pid := range sleeperPids {
		checkSleeperlResult(t, ts, pid)
	}

	for _, pid := range monPids {
		ts.Evict(pid)
	}

	ts.s.Shutdown(ts.FsLib)
}
