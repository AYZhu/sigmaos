package procclnt_test

import (
	"fmt"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	db "sigmaos/debug"
	"sigmaos/fslib"
	"sigmaos/groupmgr"
	"sigmaos/linuxsched"
	"sigmaos/perf"
	"sigmaos/proc"
	sp "sigmaos/sigmap"
	"sigmaos/test"
)

const (
	SLEEP_MSECS = 2000
)

const program = "procclnt_test"

func procd(ts *test.Tstate) string {
	st, err := ts.GetDir("name/procd")
	assert.Nil(ts.T, err, "Readdir")
	return st[0].Name
}

func spawnSpinner(t *testing.T, ts *test.Tstate) proc.Tpid {
	return spawnSpinnerNcore(ts, 0)
}

func spawnSpinnerNcore(ts *test.Tstate, ncore proc.Tcore) proc.Tpid {
	pid := proc.GenPid()
	a := proc.MakeProcPid(pid, "spinner", []string{"name/"})
	a.SetNcore(ncore)
	err := ts.Spawn(a)
	assert.Nil(ts.T, err, "Spawn")
	return pid
}

func burstSpawnSpinner(t *testing.T, ts *test.Tstate, N uint) []*proc.Proc {
	ps := make([]*proc.Proc, 0, N)
	for i := uint(0); i < N; i++ {
		p := proc.MakeProc("spinner", []string{"name/"})
		p.SetNcore(1)
		ps = append(ps, p)
	}
	failed, errs := ts.SpawnBurst(ps)
	assert.Equal(t, 0, len(failed), "Failed spawning some procs: %v", errs)
	return ps
}

func spawnSleeperWithPid(t *testing.T, ts *test.Tstate, pid proc.Tpid) {
	spawnSleeperNcore(t, ts, pid, 0, SLEEP_MSECS)
}

func spawnSleeper(t *testing.T, ts *test.Tstate) proc.Tpid {
	pid := proc.GenPid()
	spawnSleeperWithPid(t, ts, pid)
	return pid
}

func spawnSleeperNcore(t *testing.T, ts *test.Tstate, pid proc.Tpid, ncore proc.Tcore, msecs int) {
	a := proc.MakeProcPid(pid, "sleeper", []string{fmt.Sprintf("%dms", msecs), "name/"})
	a.SetNcore(ncore)
	err := ts.Spawn(a)
	assert.Nil(t, err, "Spawn")
}

func spawnSpawner(t *testing.T, ts *test.Tstate, childPid proc.Tpid, msecs int) proc.Tpid {
	p := proc.MakeProc("spawner", []string{"false", childPid.String(), "sleeper", fmt.Sprintf("%dms", msecs), "name/"})
	err := ts.Spawn(p)
	assert.Nil(t, err, "Spawn")
	return p.GetPid()
}

func checkSleeperResult(t *testing.T, ts *test.Tstate, pid proc.Tpid) bool {
	res := true
	b, err := ts.GetFile("name/" + pid.String() + "_out")
	res = assert.Nil(t, err, "GetFile") && res
	res = assert.Equal(t, string(b), "hello", "Output") && res

	return res
}

func checkSleeperResultFalse(t *testing.T, ts *test.Tstate, pid proc.Tpid) {
	b, err := ts.GetFile("name/" + pid.String() + "_out")
	assert.NotNil(t, err, "GetFile")
	assert.NotEqual(t, string(b), "hello", "Output")
}

func TestWaitExitSimpleSingle(t *testing.T) {
	ts := test.MakeTstateAll(t)

	a := proc.MakeProc("sleeper", []string{fmt.Sprintf("%dms", SLEEP_MSECS), "name/"})
	db.DPrintf(db.TEST, "Pre spawn")
	err := ts.Spawn(a)
	assert.Nil(t, err, "Spawn")
	db.DPrintf(db.TEST, "Post spawn")

	db.DPrintf(db.TEST, "Pre waitexit")
	status, err := ts.WaitExit(a.GetPid())
	db.DPrintf(db.TEST, "Post waitexit")
	assert.Nil(t, err, "WaitExit error")
	assert.True(t, status.IsStatusOK(), "Exit status wrong")

	ts.Shutdown()
}

func TestWaitExitSimpleMultiKernel(t *testing.T) {
	ts := test.MakeTstateAll(t)

	err := ts.BootNode(1)
	assert.Nil(t, err, "Boot node: %v", err)

	a := proc.MakeProc("sleeper", []string{fmt.Sprintf("%dms", SLEEP_MSECS), "name/"})
	db.DPrintf(db.TEST, "Pre spawn")
	err = ts.Spawn(a)
	db.DPrintf(db.TEST, "Post spawn")
	assert.Nil(t, err, "Spawn")

	db.DPrintf(db.TEST, "Pre waitexit")
	status, err := ts.WaitExit(a.GetPid())
	db.DPrintf(db.TEST, "Post waitexit")
	assert.Nil(t, err, "WaitExit error")
	assert.True(t, status.IsStatusOK(), "Exit status wrong")

	ts.Shutdown()
}

func TestWaitExitOne(t *testing.T) {
	ts := test.MakeTstateAll(t)

	start := time.Now()

	pid := spawnSleeper(t, ts)
	status, err := ts.WaitExit(pid)
	assert.Nil(t, err, "WaitExit error")
	assert.True(t, status.IsStatusOK(), "Exit status wrong")

	// cleaned up (may take a bit)
	time.Sleep(500 * time.Millisecond)
	_, err = ts.Stat(path.Join(sp.PROCD, "~local", proc.PIDS, pid.String()))
	assert.NotNil(t, err, "Stat %v", path.Join(proc.PIDS, pid.String()))

	end := time.Now()

	assert.True(t, end.Sub(start) > SLEEP_MSECS*time.Millisecond)

	checkSleeperResult(t, ts, pid)

	ts.Shutdown()
}

func TestWaitExitN(t *testing.T) {
	ts := test.MakeTstateAll(t)
	nProcs := 100
	var done sync.WaitGroup
	done.Add(nProcs)

	for i := 0; i < nProcs; i++ {
		go func() {
			pid := spawnSleeper(t, ts)
			status, err := ts.WaitExit(pid)
			assert.Nil(t, err, "WaitExit error")
			assert.True(t, status.IsStatusOK(), "Exit status wrong %v", status)
			db.DPrintf(db.TEST, "Exited %v", pid)

			// cleaned up (may take a bit)
			time.Sleep(500 * time.Millisecond)
			_, err = ts.Stat(path.Join(sp.PROCD, "~local", proc.PIDS, pid.String()))
			assert.NotNil(t, err, "Stat %v", path.Join(proc.PIDS, pid.String()))

			checkSleeperResult(t, ts, pid)

			done.Done()
		}()
	}
	done.Wait()
	ts.Shutdown()
}

func TestWaitExitParentRetStat(t *testing.T) {
	ts := test.MakeTstateAll(t)

	start := time.Now()

	pid := spawnSleeper(t, ts)
	time.Sleep(2 * SLEEP_MSECS * time.Millisecond)
	status, err := ts.WaitExit(pid)
	assert.Nil(t, err, "WaitExit error")
	assert.True(t, status.IsStatusOK(), "Exit status wrong")

	// cleaned up
	_, err = ts.Stat(path.Join(sp.PROCD, "~local", proc.PIDS, pid.String()))
	assert.NotNil(t, err, "Stat %v", path.Join(sp.PROCD, "~local", proc.PIDS, pid.String()))

	end := time.Now()

	assert.True(t, end.Sub(start) > SLEEP_MSECS*time.Millisecond)

	checkSleeperResult(t, ts, pid)

	ts.Shutdown()
}

func TestWaitExitParentAbandons(t *testing.T) {
	ts := test.MakeTstateAll(t)

	start := time.Now()

	cPid := proc.GenPid()
	pid := spawnSpawner(t, ts, cPid, SLEEP_MSECS)
	err := ts.WaitStart(pid)
	assert.Nil(t, err, "WaitStart error")
	status, err := ts.WaitExit(pid)
	assert.True(t, status.IsStatusOK(), "WaitExit status error")
	assert.Nil(t, err, "WaitExit error")
	// Wait for the child to run & finish
	time.Sleep(2 * SLEEP_MSECS * time.Millisecond)

	// cleaned up
	_, err = ts.Stat(path.Join(sp.PROCD, "~local", proc.PIDS, pid.String()))
	assert.NotNil(t, err, "Stat")

	end := time.Now()

	assert.True(t, end.Sub(start) > SLEEP_MSECS*time.Millisecond)

	checkSleeperResult(t, ts, cPid)

	ts.Shutdown()
}

func TestWaitStart(t *testing.T) {
	ts := test.MakeTstateAll(t)

	start := time.Now()

	pid := spawnSleeper(t, ts)
	err := ts.WaitStart(pid)
	assert.Nil(t, err, "WaitStart error")

	end := time.Now()

	assert.True(t, end.Sub(start) < SLEEP_MSECS*time.Millisecond, "WaitStart waited too long")

	// Check if proc exists
	sts, err := ts.GetDir(path.Join("name/procd", procd(ts), sp.PROCD_RUNNING))
	assert.Nil(t, err, "Readdir")
	assert.True(t, fslib.Present(sts, []string{pid.String()}), "pid")

	// Make sure the proc hasn't finished yet...
	checkSleeperResultFalse(t, ts, pid)

	ts.WaitExit(pid)

	checkSleeperResult(t, ts, pid)

	ts.Shutdown()
}

// Should exit immediately
func TestWaitNonexistentProc(t *testing.T) {
	ts := test.MakeTstateAll(t)

	ch := make(chan bool)

	pid := proc.GenPid()
	go func() {
		ts.WaitExit(pid)
		ch <- true
	}()

	done := <-ch
	assert.True(t, done, "Nonexistent proc")

	close(ch)

	ts.Shutdown()
}

func TestSpawnManyProcsParallel(t *testing.T) {
	ts := test.MakeTstateAll(t)

	const N_CONCUR = 13
	const N_SPAWNS = 500

	err := ts.BootNode(1)
	assert.Nil(t, err, "BootProcd 1")

	err = ts.BootNode(1)
	assert.Nil(t, err, "BootProcd 2")

	done := make(chan int)

	for i := 0; i < N_CONCUR; i++ {
		go func(i int) {
			for j := 0; j < N_SPAWNS; j++ {
				pid := proc.GenPid()
				db.DPrintf(db.TEST, "Prep spawn %v", pid)
				a := proc.MakeProcPid(pid, "sleeper", []string{"0ms", "name/"})
				_, errs := ts.SpawnBurst([]*proc.Proc{a})
				assert.True(t, len(errs) == 0, "Spawn err %v", errs)
				db.DPrintf(db.TEST, "Done spawn %v", pid)

				db.DPrintf(db.TEST, "Prep WaitStart %v", pid)
				err := ts.WaitStart(a.GetPid())
				db.DPrintf(db.TEST, "Done WaitStart %v", pid)
				assert.Nil(t, err, "WaitStart error")

				db.DPrintf(db.TEST, "Prep WaitExit %v", pid)
				status, err := ts.WaitExit(a.GetPid())
				db.DPrintf(db.TEST, "Done WaitExit %v", pid)
				assert.Nil(t, err, "WaitExit")
				assert.True(t, status.IsStatusOK(), "Status not OK")
			}
			done <- i
		}(i)
	}
	for i := 0; i < N_CONCUR; i++ {
		x := <-done
		db.DPrintf(db.TEST, "Done %v", x)
	}

	ts.Shutdown()
}

func TestCrashProcOne(t *testing.T) {
	ts := test.MakeTstateAll(t)

	a := proc.MakeProc("crash", []string{})
	err := ts.Spawn(a)
	assert.Nil(t, err, "Spawn")

	err = ts.WaitStart(a.GetPid())
	assert.Nil(t, err, "WaitStart error")

	status, err := ts.WaitExit(a.GetPid())
	assert.Nil(t, err, "WaitExit")
	assert.True(t, status.IsStatusErr(), "Status not err")
	assert.Equal(t, "exit status 2", status.Msg(), "WaitExit")

	ts.Shutdown()
}

func TestEarlyExit1(t *testing.T) {
	ts := test.MakeTstateAll(t)

	pid1 := proc.GenPid()
	a := proc.MakeProc("parentexit", []string{fmt.Sprintf("%dms", SLEEP_MSECS), pid1.String()})
	err := ts.Spawn(a)
	assert.Nil(t, err, "Spawn")

	// Wait for parent to finish
	status, err := ts.WaitExit(a.GetPid())
	assert.Nil(t, err, "WaitExit")
	assert.True(t, status.IsStatusOK(), "WaitExit")

	// Child should not have terminated yet.
	checkSleeperResultFalse(t, ts, pid1)

	time.Sleep(2 * SLEEP_MSECS * time.Millisecond)

	// Child should have exited
	b, err := ts.GetFile("name/" + pid1.String() + "_out")
	assert.Nil(t, err, "GetFile")
	assert.Equal(t, string(b), "hello", "Output")

	// .. and cleaned up
	_, err = ts.Stat(path.Join(sp.PROCD, "~local", proc.PIDS, pid1.String()))
	assert.NotNil(t, err, "Stat")

	ts.Shutdown()
}

func TestEarlyExitN(t *testing.T) {
	ts := test.MakeTstateAll(t)
	nProcs := 500
	var done sync.WaitGroup
	done.Add(nProcs)

	for i := 0; i < nProcs; i++ {
		go func(i int) {
			pid1 := proc.GenPid()
			a := proc.MakeProc("parentexit", []string{fmt.Sprintf("%dms", 0), pid1.String()})
			err := ts.Spawn(a)
			assert.Nil(t, err, "Spawn")

			// Wait for parent to finish
			status, err := ts.WaitExit(a.GetPid())
			assert.Nil(t, err, "WaitExit err: %v", err)
			assert.True(t, status.IsStatusOK(), "WaitExit: %v", status)

			time.Sleep(2 * SLEEP_MSECS * time.Millisecond)

			// Child should have exited
			b, err := ts.GetFile("name/" + pid1.String() + "_out")
			assert.Nil(t, err, "GetFile")
			assert.Equal(t, string(b), "hello", "Output")

			// .. and cleaned up
			_, err = ts.Stat(path.Join(sp.PROCD, "~local", proc.PIDS, pid1.String()))
			assert.NotNil(t, err, "Stat")
			done.Done()
		}(i)
	}
	done.Wait()

	ts.Shutdown()
}

// Spawn a bunch of procs concurrently, then wait for all of them & check
// their result
func TestConcurrentProcs(t *testing.T) {
	ts := test.MakeTstateAll(t)

	nProcs := 8
	pids := map[proc.Tpid]int{}

	var barrier sync.WaitGroup
	barrier.Add(nProcs)
	var started sync.WaitGroup
	started.Add(nProcs)
	var done sync.WaitGroup
	done.Add(nProcs)

	for i := 0; i < nProcs; i++ {
		pid := proc.GenPid()
		_, alreadySpawned := pids[pid]
		for alreadySpawned {
			pid = proc.GenPid()
			_, alreadySpawned = pids[pid]
		}
		pids[pid] = i
		go func(pid proc.Tpid, started *sync.WaitGroup, i int) {
			barrier.Done()
			barrier.Wait()
			spawnSleeperWithPid(t, ts, pid)
			started.Done()
		}(pid, &started, i)
	}

	started.Wait()

	for pid, i := range pids {
		_ = i
		go func(pid proc.Tpid, done *sync.WaitGroup, i int) {
			defer done.Done()
			ts.WaitExit(pid)
			checkSleeperResult(t, ts, pid)
			time.Sleep(100 * time.Millisecond)
			_, err := ts.Stat(path.Join(sp.PROCD, "~local", proc.PIDS, pid.String()))
			assert.NotNil(t, err, "Stat %v", path.Join(proc.PIDS, pid.String()))
		}(pid, &done, i)
	}

	done.Wait()

	ts.Shutdown()
}

func evict(ts *test.Tstate, pid proc.Tpid) {
	err := ts.WaitStart(pid)
	assert.Nil(ts.T, err, "Wait start err %v", err)
	time.Sleep(SLEEP_MSECS * time.Millisecond)
	err = ts.Evict(pid)
	assert.Nil(ts.T, err, "evict")
}

func TestEvict(t *testing.T) {
	ts := test.MakeTstateAll(t)

	pid := spawnSpinner(t, ts)

	go evict(ts, pid)

	status, err := ts.WaitExit(pid)
	assert.Nil(t, err, "WaitExit")
	assert.True(t, status.IsStatusEvicted(), "WaitExit status")

	ts.Shutdown()
}

func TestReserveCores(t *testing.T) {
	ts := test.MakeTstateAll(t)

	start := time.Now()
	pid := proc.Tpid("sleeper-aaaaaaa")
	spawnSleeperNcore(t, ts, pid, proc.Tcore(linuxsched.NCores), SLEEP_MSECS)

	// Make sure pid1 is alphabetically sorted after pid, to ensure that this
	// proc is only picked up *after* the other one.
	pid1 := proc.Tpid("sleeper-bbbbbb")
	spawnSleeperNcore(t, ts, pid1, 1, SLEEP_MSECS)

	status, err := ts.WaitExit(pid)
	assert.Nil(t, err, "WaitExit")
	assert.True(t, status.IsStatusOK(), "WaitExit status")

	// Make sure the second proc didn't finish
	checkSleeperResult(t, ts, pid)
	checkSleeperResultFalse(t, ts, pid1)

	status, err = ts.WaitExit(pid1)
	assert.Nil(t, err, "WaitExit 2")
	assert.True(t, status.IsStatusOK(), "WaitExit status 2")
	end := time.Now()

	assert.True(t, end.Sub(start) > (SLEEP_MSECS*2)*time.Millisecond, "Parallelized")

	ts.Shutdown()
}

func TestWorkStealing(t *testing.T) {
	assert.True(t, false, "WorkStealing not implemented")
	return

	ts := test.MakeTstateAll(t)

	err := ts.BootNode(1)
	assert.Nil(t, err, "Boot node %v", err)

	pid := spawnSpinnerNcore(ts, proc.Tcore(linuxsched.NCores))

	pid1 := spawnSpinnerNcore(ts, proc.Tcore(linuxsched.NCores))

	err = ts.WaitStart(pid)
	assert.Nil(t, err, "WaitStart")

	err = ts.WaitStart(pid1)
	assert.Nil(t, err, "WaitStart")

	err = ts.Evict(pid)
	assert.Nil(t, err, "Evict")

	err = ts.Evict(pid1)
	assert.Nil(t, err, "Evict")

	status, err := ts.WaitExit(pid)
	assert.Nil(t, err, "WaitExit")
	assert.True(t, status.IsStatusEvicted(), "WaitExit status")

	status, err = ts.WaitExit(pid1)
	assert.Nil(t, err, "WaitExit 2")
	assert.True(t, status.IsStatusEvicted(), "WaitExit status 2")

	// Check that work-stealing symlinks were cleaned up.
	sts, _, err := ts.ReadDir(path.Join(sp.PROCD_WS, sp.PROCD_RUNQ_LC))
	assert.Nil(t, err, "Readdir %v", err)
	assert.Equal(t, 0, len(sts), "Wrong length ws dir[%v]: %v", path.Join(sp.PROCD_WS, sp.PROCD_RUNQ_LC), sts)

	sts, _, err = ts.ReadDir(path.Join(sp.PROCD_WS, sp.PROCD_RUNQ_BE))
	assert.Nil(t, err, "Readdir %v", err)
	assert.Equal(t, 0, len(sts), "Wrong length ws dir[%v]: %v", path.Join(sp.PROCD_WS, sp.PROCD_RUNQ_BE), sts)

	ts.Shutdown()
}

func TestEvictN(t *testing.T) {
	ts := test.MakeTstateAll(t)

	N := int(linuxsched.NCores)

	pids := []proc.Tpid{}
	for i := 0; i < N; i++ {
		pid := spawnSpinner(t, ts)
		pids = append(pids, pid)
		go evict(ts, pid)
	}

	for i := 0; i < N; i++ {
		status, err := ts.WaitExit(pids[i])
		assert.Nil(t, err, "WaitExit")
		assert.True(t, status.IsStatusEvicted(), "WaitExit status")
	}

	ts.Shutdown()
}

func getNChildren(ts *test.Tstate) int {
	c, err := ts.GetChildren()
	assert.Nil(ts.T, err, "getnchildren")
	return len(c)
}

func TestBurstSpawn(t *testing.T) {
	assert.True(t, false, "Burst Spawn not implemented")
	return

	ts := test.MakeTstateAll(t)

	// Number of spinners to burst-spawn
	N := linuxsched.NCores * 3

	// Start a couple new procds.
	err := ts.BootNode(1)
	assert.Nil(t, err, "BootNode %v", err)
	err = ts.BootNode(1)
	assert.Nil(t, err, "BootNode %v", err)

	ps := burstSpawnSpinner(t, ts, N)

	for _, p := range ps {
		err := ts.WaitStart(p.GetPid())
		assert.Nil(t, err, "WaitStart: %v", err)
	}

	for _, p := range ps {
		err := ts.Evict(p.GetPid())
		assert.Nil(t, err, "Evict: %v", err)
	}

	for _, p := range ps {
		status, err := ts.WaitExit(p.GetPid())
		assert.Nil(t, err, "WaitExit: %v", err)
		assert.True(t, status.IsStatusEvicted(), "Wrong status: %v", status)
	}

	ts.Shutdown()
}

func TestSpawnProcdCrash(t *testing.T) {
	assert.True(t, false, "Crash not implemented")
	return

	ts := test.MakeTstateAll(t)

	// Spawn a proc which can't possibly be run by any procd.
	pid := spawnSpinnerNcore(ts, proc.Tcore(linuxsched.NCores*2))

	assert.True(t, false, "KillOne")
	_ = pid

	//	err := ts.KillOne(sp.PROCDREL)
	//	assert.Nil(t, err, "KillOne: %v", err)
	//
	//	err = ts.WaitStart(pid)
	//	assert.NotNil(t, err, "WaitStart: %v", err)
	//
	//	_, err = ts.WaitExit(pid)
	//	assert.NotNil(t, err, "WaitExit: %v", err)

	ts.Shutdown()
}

func TestMaintainReplicationLevelCrashProcd(t *testing.T) {
	assert.True(t, false, "Crash not implemented")
	return

	ts := test.MakeTstateAll(t)

	N_REPL := 3
	OUTDIR := "name/spinner-ephs"

	// Start a couple new procds.
	err := ts.BootNode(1)
	assert.Nil(t, err, "BootNode %v", err)
	err = ts.BootNode(1)
	assert.Nil(t, err, "BootNode %v", err)

	// Count number of children.
	nChildren := getNChildren(ts)

	err = ts.MkDir(OUTDIR, 0777)
	assert.Nil(t, err, "Mkdir")

	// Start a bunch of replicated spinner procs.
	sm := groupmgr.Start(ts.FsLib, ts.ProcClnt, N_REPL, "spinner", []string{}, OUTDIR, 0, N_REPL, 0, 0, 0)
	nChildren += N_REPL

	// Wait for them to spawn.
	time.Sleep(1 * time.Second)

	// Make sure they spawned correctly.
	st, err := ts.GetDir(OUTDIR)
	assert.Nil(t, err, "readdir1")
	assert.Equal(t, N_REPL, len(st), "wrong num spinners check #1")
	assert.Equal(t, nChildren, getNChildren(ts), "wrong num children")

	assert.True(t, false, "KillOne")
	_ = sm
	//	err = ts.KillOne(sp.PROCDREL)
	//	assert.Nil(t, err, "kill procd")
	//
	//	// Wait for them to respawn.
	//	time.Sleep(5 * time.Second)
	//
	//	// Make sure they spawned correctly.
	//	st, err = ts.GetDir(OUTDIR)
	//	assert.Nil(t, err, "readdir1")
	//	assert.Equal(t, N_REPL, len(st), "wrong num spinners check #2")
	//
	//	err = ts.KillOne(sp.PROCDREL)
	//	assert.Nil(t, err, "kill procd")
	//
	//	// Wait for them to respawn.
	//	time.Sleep(5 * time.Second)
	//
	//	// Make sure they spawned correctly.
	//	st, err = ts.GetDir(OUTDIR)
	//	assert.Nil(t, err, "readdir1")
	//	assert.Equal(t, N_REPL, len(st), "wrong num spinners check #3")
	//
	//	sm.Stop()

	ts.Shutdown()
}

// Test to see if any core has a spinner running on it (high utilization).
func anyCoresOccupied(coresMaps []map[string]bool) bool {
	N_SAMPLES := 5
	// Calculate the average utilization over a 250ms period for each core to be
	// revoked.
	coreOccupied := false
	for c, m := range coresMaps {
		idle0, total0 := perf.GetCPUSample(m)
		idleDelta := uint64(0)
		totalDelta := uint64(0)
		// Collect some CPU util samples for this core.
		for i := 0; i < N_SAMPLES; i++ {
			time.Sleep(25 * time.Millisecond)
			idle1, total1 := perf.GetCPUSample(m)
			idleDelta += idle1 - idle0
			totalDelta += total1 - total0
			idle0 = idle1
			total0 = total1
		}
		avgCoreUtil := 100.0 * ((float64(totalDelta) - float64(idleDelta)) / float64(totalDelta))
		db.DPrintf(db.TEST, "Core %v utilization: %v", c, avgCoreUtil)
		if avgCoreUtil > 50.0 {
			coreOccupied = true
		}
	}
	return coreOccupied
}

//func TestProcdResize1(t *testing.T) {
//	ts := test.MakeTstateAll(t)
//
//	// Run a proc that claims all cores.
//	pid := proc.GenPid()
//	spawnSleeperNcore(t, ts, pid, proc.Tcore(linuxsched.NCores), SLEEP_MSECS)
//	status, err := ts.WaitExit(pid)
//	assert.Nil(t, err, "WaitExit")
//	assert.True(t, status.IsStatusOK(), "WaitExit status")
//	checkSleeperResult(t, ts, pid)
//
//	nCoresToRevoke := int(math.Ceil(float64(linuxsched.NCores)/2 + 1))
//	coreIv := sessp.MkInterval(0, uint64(nCoresToRevoke))
//
//	ctlFilePath := path.Join(sp.PROCD, "~local", sp.RESOURCE_CTL)
//
//	// Remove some cores from the procd.
//	db.DPrintf(db.TEST, "Removing %v cores %v from procd.", nCoresToRevoke, coreIv)
//	revokeMsg := resource.MakeResourceMsg(resource.Trequest, resource.Tcore, coreIv.Marshal(), nCoresToRevoke)
//	_, err = ts.SetFile(ctlFilePath, revokeMsg.Marshal(), sp.OWRITE, 0)
//	assert.Nil(t, err, "SetFile revoke: %v", err)
//
//	// Run a proc which shouldn't fit on the newly resized procd.
//	db.DPrintf(db.TEST, "Spawning a proc which shouldn't fit.")
//	pid1 := proc.GenPid()
//	spawnSleeperNcore(t, ts, pid1, proc.Tcore(linuxsched.NCores), SLEEP_MSECS)
//
//	time.Sleep(3 * SLEEP_MSECS)
//	// Proc should not have run.
//	checkSleeperResultFalse(t, ts, pid1)
//
//	pid2 := proc.GenPid()
//	db.DPrintf(db.TEST, "Spawning a proc which should fit.")
//	spawnSleeperNcore(t, ts, pid2, proc.Tcore(linuxsched.NCores/2-1), SLEEP_MSECS)
//	status, err = ts.WaitExit(pid2)
//	assert.Nil(t, err, "WaitExit 2")
//	assert.True(t, status.IsStatusOK(), "WaitExit status 2")
//	checkSleeperResult(t, ts, pid2)
//	db.DPrintf(db.TEST, "Proc which should fit ran")
//
//	// Grant the procd back its cores.
//	db.DPrintf(db.TEST, "Granting %v cores %v to procd.", nCoresToRevoke, coreIv)
//	grantMsg := resource.MakeResourceMsg(resource.Tgrant, resource.Tcore, coreIv.Marshal(), nCoresToRevoke)
//	_, err = ts.SetFile(ctlFilePath, grantMsg.Marshal(), sp.OWRITE, 0)
//	assert.Nil(t, err, "SetFile grant: %v", err)
//
//	// Make sure the proc ran.
//	status, err = ts.WaitExit(pid1)
//	assert.Nil(t, err, "WaitExit 3")
//	assert.True(t, status.IsStatusOK(), "WaitExit status 3")
//	checkSleeperResult(t, ts, pid1)
//
//	ts.Shutdown()
//}
//
//func TestProcdResizeN(t *testing.T) {
//	ts := test.MakeTstateAll(t)
//
//	N := 5
//
//	nCoresToRevoke := int(math.Ceil(float64(linuxsched.NCores)/2 + 1))
//	coreIv := sessp.MkInterval(0, uint64(nCoresToRevoke))
//
//	ctlFilePath := path.Join(sp.PROCD, "~local", sp.RESOURCE_CTL)
//	for i := 0; i < N; i++ {
//		db.DPrintf(db.TEST, "Resize i=%v", i)
//		// Run a proc that claims all cores.
//		pid := proc.GenPid()
//		spawnSleeperNcore(t, ts, pid, proc.Tcore(linuxsched.NCores), SLEEP_MSECS)
//		status, err := ts.WaitExit(pid)
//		assert.Nil(t, err, "WaitExit")
//		assert.True(t, status.IsStatusOK(), "WaitExit status")
//		checkSleeperResult(t, ts, pid)
//
//		// Remove some cores from the procd.
//		db.DPrintf(db.TEST, "Removing %v cores %v from procd.", nCoresToRevoke, coreIv)
//		revokeMsg := resource.MakeResourceMsg(resource.Trequest, resource.Tcore, coreIv.Marshal(), nCoresToRevoke)
//		_, err = ts.SetFile(ctlFilePath, revokeMsg.Marshal(), sp.OWRITE, 0)
//		assert.Nil(t, err, "SetFile revoke: %v", err)
//
//		// Run a proc which shouldn't fit on the newly resized procd.
//		db.DPrintf(db.TEST, "Spawning a proc which shouldn't fit.")
//		pid1 := proc.GenPid()
//		spawnSleeperNcore(t, ts, pid1, proc.Tcore(linuxsched.NCores), SLEEP_MSECS)
//
//		time.Sleep(3 * SLEEP_MSECS)
//		// Proc should not have run.
//		checkSleeperResultFalse(t, ts, pid1)
//
//		pid2 := proc.GenPid()
//		db.DPrintf(db.TEST, "Spawning a proc which should fit.")
//		spawnSleeperNcore(t, ts, pid2, proc.Tcore(linuxsched.NCores/2-1), SLEEP_MSECS)
//		status, err = ts.WaitExit(pid2)
//		assert.Nil(t, err, "WaitExit 2")
//		assert.True(t, status.IsStatusOK(), "WaitExit status 2")
//		checkSleeperResult(t, ts, pid2)
//		db.DPrintf(db.TEST, "Proc which should fit ran")
//
//		// Grant the procd back its cores.
//		db.DPrintf(db.TEST, "Granting %v cores %v to procd.", nCoresToRevoke, coreIv)
//		grantMsg := resource.MakeResourceMsg(resource.Tgrant, resource.Tcore, coreIv.Marshal(), nCoresToRevoke)
//		_, err = ts.SetFile(ctlFilePath, grantMsg.Marshal(), sp.OWRITE, 0)
//		assert.Nil(t, err, "SetFile grant: %v", err)
//
//		// Make sure the proc ran.
//		status, err = ts.WaitExit(pid1)
//		assert.Nil(t, err, "WaitExit 3")
//		assert.True(t, status.IsStatusOK(), "WaitExit status 3")
//		checkSleeperResult(t, ts, pid1)
//	}
//
//	ts.Shutdown()
//}
//
//func TestProcdResizeAccurateStats(t *testing.T) {
//	ts := test.MakeTstateAll(t)
//
//	// Spawn NCores/2 spinners, each claiming two cores.
//	pids := []proc.Tpid{}
//	for i := 0; i < int(linuxsched.NCores)/2; i++ {
//		pid := spawnSpinnerNcore(ts, proc.Tcore(2))
//		err := ts.WaitStart(pid)
//		assert.Nil(t, err, "WaitStart")
//		pids = append(pids, pid)
//	}
//
//	// Revoke half of the procd's cores.
//	nCoresToRevoke := int(math.Ceil(float64(linuxsched.NCores) / 2))
//	coreIv := sessp.MkInterval(0, uint64(nCoresToRevoke))
//
//	ctlFilePath := path.Join(sp.PROCD, "~local", sp.RESOURCE_CTL)
//
//	// Remove some cores from the procd.
//	db.DPrintf(db.TEST, "Removing %v cores %v from procd.", nCoresToRevoke, coreIv)
//	revokeMsg := resource.MakeResourceMsg(resource.Trequest, resource.Tcore, coreIv.Marshal(), nCoresToRevoke)
//	_, err := ts.SetFile(ctlFilePath, revokeMsg.Marshal(), sp.OWRITE, 0)
//	assert.Nil(t, err, "SetFile revoke: %v", err)
//
//	// Sleep for a bit
//	time.Sleep(SLEEP_MSECS * time.Millisecond)
//
//	// Get the procd's utilization.
//	st := stats.StatInfo{}
//	err = ts.GetFileJson(path.Join(sp.PROCD, "~local", sp.STATSD), &st)
//	assert.Nil(t, err, "statsd: %v", err)
//
//	// Ensure that the procd is accurately representing the utilization (it
//	// should show ~100% CPU utilization, not 50%).
//	db.DPrintf(db.TEST, "Stats after shrink: %v", st)
//	assert.True(t, st.Util > 90.0, "Util too low, %v < 90", st.Util)
//
//	// Grant the procd back its cores.
//	db.DPrintf(db.TEST, "Granting %v cores %v to procd.", nCoresToRevoke, coreIv)
//	grantMsg := resource.MakeResourceMsg(resource.Tgrant, resource.Tcore, coreIv.Marshal(), nCoresToRevoke)
//	_, err = ts.SetFile(ctlFilePath, grantMsg.Marshal(), sp.OWRITE, 0)
//	assert.Nil(t, err, "SetFile grant: %v", err)
//
//	// Sleep for a bit
//	time.Sleep(SLEEP_MSECS * time.Millisecond)
//
//	// Get the procd's utilization again.
//	err = ts.GetFileJson(path.Join(sp.PROCD, "~local", sp.STATSD), &st)
//	assert.Nil(t, err, "statsd: %v", err)
//
//	// Ensure that the procd's utilization has been adjusted again (it
//	// should show ~50% CPU utilization, not 100%).
//	db.DPrintf(db.TEST, "Stats after shrink: %v", st)
//	assert.True(t, st.Util < 60.0, "Util too high, %v > 60", st.Util)
//
//	// Evict all of the spinning procs.
//	for _, pid := range pids {
//		err := ts.Evict(pid)
//		assert.Nil(ts.T, err, "Evict")
//		status, err := ts.WaitExit(pid)
//		assert.Nil(t, err, "WaitExit")
//		assert.True(t, status.IsStatusEvicted(), "WaitExit status")
//	}
//
//	ts.Shutdown()
//}
//func TestProcdResizeCoreRepinning(t *testing.T) {
//	ts := test.MakeTstateAll(t)
//
//	// Spawn NCores/2 spinners, each claiming two cores.
//	pids := []proc.Tpid{}
//	for i := 0; i < int(linuxsched.NCores)/2; i++ {
//		pid := spawnSpinnerNcore(ts, proc.Tcore(2))
//		err := ts.WaitStart(pid)
//		assert.Nil(t, err, "WaitStart")
//		pids = append(pids, pid)
//	}
//
//	// Revoke half of the procd's cores.
//	nCoresToRevoke := int(math.Ceil(float64(linuxsched.NCores) / 2))
//	coreIv := sessp.MkInterval(0, uint64(nCoresToRevoke))
//
//	ctlFilePath := path.Join(sp.PROCD, "~local", sp.RESOURCE_CTL)
//
//	// Create a map to sample core utilization levels on the cores which will be
//	// revoked.
//	coresMaps := []map[string]bool{}
//	for i := coreIv.Start; i < coreIv.End; i++ {
//		coresMaps = append(coresMaps, map[string]bool{"cpu" + strconv.Itoa(int(i)): true})
//	}
//
//	coreOccupied := anyCoresOccupied(coresMaps)
//	// Make sure that at least some of the cores to be revoked has a spinning
//	// proc on it.
//	assert.True(t, coreOccupied, "No cores occupied")
//
//	// Remove some cores from the procd.
//	db.DPrintf(db.TEST, "Removing %v cores %v from procd.", nCoresToRevoke, coreIv)
//	revokeMsg := resource.MakeResourceMsg(resource.Trequest, resource.Tcore, coreIv.Marshal(), nCoresToRevoke)
//	_, err := ts.SetFile(ctlFilePath, revokeMsg.Marshal(), sp.OWRITE, 0)
//	assert.Nil(t, err, "SetFile revoke: %v", err)
//
//	// Sleep for a bit
//	time.Sleep(SLEEP_MSECS * time.Millisecond)
//
//	coreOccupied = anyCoresOccupied(coresMaps)
//	// Ensure that none of the revoked cores have spinning procs running on them.
//	assert.False(t, coreOccupied, "Core still occupied")
//
//	// Evict all of the spinning procs.
//	for _, pid := range pids {
//		err := ts.Evict(pid)
//		assert.Nil(ts.T, err, "Evict")
//		status, err := ts.WaitExit(pid)
//		assert.Nil(t, err, "WaitExit")
//		assert.True(t, status.IsStatusEvicted(), "WaitExit status")
//	}
//
//	ts.Shutdown()
//}
