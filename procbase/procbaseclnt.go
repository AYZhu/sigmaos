package procbase

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"strings"

	"github.com/thanhpk/randstr"

	"ulambda/fslib"
	"ulambda/named"
	"ulambda/namespace"
	np "ulambda/ninep"
	"ulambda/proc"
	"ulambda/seccomp"
	"ulambda/sync"
)

type Twait uint32

const (
	START Twait = 0
	EXIT  Twait = 1
)

const (
	RUNQ = "name/runq"
)

const (
	START_COND      = "start-cond."
	EVICT_COND      = "evict-cond."
	EXIT_COND       = "exit-cond."
	RET_STAT        = "ret-stat."
	PARENT_RET_STAT = "parent-ret-stat."
)

const (
	RUNQLC_PRIORITY = "0"
	RUNQ_PRIORITY   = "1"
)

type ProcBaseClnt struct {
	runq *sync.FilePriorityBag
	*fslib.FsLib
}

func MakeProcBaseClnt(fsl *fslib.FsLib) *ProcBaseClnt {
	clnt := &ProcBaseClnt{}
	clnt.runq = sync.MakeFilePriorityBag(fsl, RUNQ)
	clnt.FsLib = fsl

	return clnt
}

// ========== SPAWN ==========

func (clnt *ProcBaseClnt) Spawn(gp proc.GenericProc) error {
	p := gp.GetProc()

	pStartCond := sync.MakeCond(clnt.FsLib, path.Join(named.PROC_COND, START_COND+p.Pid), nil, true)
	pStartCond.Init()

	pExitCond := sync.MakeCond(clnt.FsLib, path.Join(named.PROC_COND, EXIT_COND+p.Pid), nil, true)
	pExitCond.Init()

	pEvictCond := sync.MakeCond(clnt.FsLib, path.Join(named.PROC_COND, EVICT_COND+p.Pid), nil, true)
	pEvictCond.Init()

	clnt.makeParentRetStatFile(p.Pid)

	clnt.makeRetStatWaiterFile(p.Pid)

	b, err := json.Marshal(p)
	if err != nil {
		// Unlock the waiter file if unmarshal failed
		pStartCond.Destroy()
		pExitCond.Destroy()
		pEvictCond.Destroy()
		log.Fatalf("Error marshal: %v", err)
		return err
	}

	err = clnt.WriteFile(path.Join("name/procd/~ip", named.PROC_CTL_FILE), b)

	return nil
}

// ========== WAIT ==========

// Wait until a proc has started. If the proc doesn't exist, return immediately.
func (clnt *ProcBaseClnt) WaitStart(pid string) error {
	pStartCond := sync.MakeCond(clnt.FsLib, path.Join(named.PROC_COND, START_COND+pid), nil, false)
	pStartCond.Wait()
	return nil
}

// Wait until a proc has exited. If the proc doesn't exist, return immediately.
func (clnt *ProcBaseClnt) WaitExit(pid string) (string, error) {
	// Register that we want to get a return status back
	fpath, _ := clnt.registerRetStatWaiter(pid)

	// Wait for the process to exit
	pExitCond := sync.MakeCond(clnt.FsLib, path.Join(named.PROC_COND, EXIT_COND+pid), nil, false)
	pExitCond.Wait()

	status := clnt.getRetStat(pid, fpath)

	return status, nil
}

// Wait for a proc's eviction notice. If the proc doesn't exist, return immediately.
func (clnt *ProcBaseClnt) WaitEvict(pid string) error {
	pEvictCond := sync.MakeCond(clnt.FsLib, path.Join(named.PROC_COND, EVICT_COND+pid), nil, false)
	pEvictCond.Wait()
	return nil
}

// ========== STARTED ==========

// Mark that a process has started.
func (clnt *ProcBaseClnt) Started(pid string) error {
	pStartCond := sync.MakeCond(clnt.FsLib, path.Join(named.PROC_COND, START_COND+pid), nil, true)
	pStartCond.Destroy()
	// Isolate the process namespace
	newRoot := os.Getenv("NEWROOT")
	if err := namespace.Isolate(newRoot); err != nil {
		log.Fatalf("Error Isolate in clnt.Started: %v", err)
	}
	// Load a seccomp filter.
	seccomp.LoadFilter()
	return nil
}

// ========== EXITED ==========

// Mark that a process has exited.
func (clnt *ProcBaseClnt) Exited(pid string, status string) error {

	// Write back return statuses
	clnt.writeBackRetStats(pid, status)

	pExitCond := sync.MakeCond(clnt.FsLib, path.Join(named.PROC_COND, EXIT_COND+pid), nil, true)
	pExitCond.Destroy()
	return nil
}

// ========== EVICT ==========

// Notify a process that it will be evicted.
func (clnt *ProcBaseClnt) Evict(pid string) error {
	pEvictCond := sync.MakeCond(clnt.FsLib, path.Join(named.PROC_COND, EVICT_COND+pid), nil, false)
	pEvictCond.Destroy()
	return nil
}

// ========== Helpers ==========

func (clnt *ProcBaseClnt) makeParentRetStatFile(pid string) {
	if err := clnt.MakeFile(path.Join(named.TMP, PARENT_RET_STAT+pid), 0777|np.DMTMP, np.OWRITE, []byte{}); err != nil && !strings.Contains(err.Error(), "Name exists") {
		log.Fatalf("Error MakeFile in ProcBaseClnt.makeParentRetStatFile: %v", err)
	}
}

type RetStatWaiters struct {
	Fpaths []string
}

func (clnt *ProcBaseClnt) makeRetStatWaiterFile(pid string) {
	l := sync.MakeLock(clnt.FsLib, named.LOCKS, RET_STAT+pid, true)
	l.Lock()
	defer l.Unlock()

	if err := clnt.MakeFileJson(path.Join(named.PROC_RET_STAT, RET_STAT+pid), 0777, &RetStatWaiters{}); err != nil && !strings.Contains(err.Error(), "Name exists") {
		log.Fatalf("Error MakeFileJson in ProcBaseClnt.makeRetStatWaiterFile: %v", err)
	}
}

// Write back exit statuses to backwards pointers
func (clnt *ProcBaseClnt) writeBackRetStats(pid string, status string) {
	l := sync.MakeLock(clnt.FsLib, named.LOCKS, RET_STAT+pid, true)
	l.Lock()
	defer l.Unlock()

	clnt.SetFile(path.Join(named.TMP, PARENT_RET_STAT+pid), []byte(status), np.NoV)

	rswPath := path.Join(named.PROC_RET_STAT, RET_STAT+pid)
	rsw := &RetStatWaiters{}
	if err := clnt.ReadFileJson(rswPath, rsw); err != nil {
		log.Fatalf("Error ReadFileJson in ProcBaseClnt.writeBackRetStats: %v", err)
	}

	for _, fpath := range rsw.Fpaths {
		if _, err := clnt.SetFile(fpath, []byte(status), np.NoV); err != nil {
			log.Fatalf("Error WriteFile in ProcBaseClnt.writeBackRetStats: %v", err)
		}
	}

	if err := clnt.Remove(rswPath); err != nil {
		log.Fatalf("Error Remove in ProcBaseClnt.writeBackRetStats: %v", err)
	}
}

// Register backwards pointers for files in which to write return statuses
func (clnt *ProcBaseClnt) registerRetStatWaiter(pid string) (string, error) {
	l := sync.MakeLock(clnt.FsLib, named.LOCKS, RET_STAT+pid, true)
	l.Lock()
	defer l.Unlock()

	// Get & update the list of backwards pointers
	rswPath := path.Join(named.PROC_RET_STAT, RET_STAT+pid)
	rsw := &RetStatWaiters{}
	if err := clnt.ReadFileJson(rswPath, rsw); err != nil {
		if err.Error() == "file not found "+RET_STAT+pid {
			return "", err
		}
		log.Fatalf("Error ReadFileJson in ProcBaseClnt.registerRetStatWaiter: %v", err)
		return "", err
	}
	fpath := path.Join(named.TMP, randstr.Hex(16))
	rsw.Fpaths = append(rsw.Fpaths, fpath)

	if err := clnt.MakeFile(fpath, 0777, np.OWRITE, []byte{}); err != nil {
		log.Fatalf("Error MakeFile in ProcBaseClnt.registerRetStatWaiter: %v", err)
	}

	if err := clnt.WriteFileJson(rswPath, rsw); err != nil {
		log.Fatalf("Error WriteFileJson in ProcBaseClnt.registerRetStatWaiter: %v", err)
		return "", err
	}

	return fpath, nil
}

// Read & destroy a return status file
func (clnt *ProcBaseClnt) getRetStat(pid string, fpath string) string {
	var b []byte
	var err error

	// If we didn't successfully register the ret stat file
	if fpath == "" {
		// Try to read the parent ret stat file
		b, _, err = clnt.GetFile(path.Join(named.TMP, PARENT_RET_STAT+pid))
	} else {
		// Try to read the registerred ret stat file
		b, _, err = clnt.GetFile(fpath)
		if err != nil {
			log.Fatalf("Error ReadFile in ProcBaseClnt.getRetStat: %v", err)
		}

		// Remove the registerred ret stat file
		if err := clnt.Remove(fpath); err != nil {
			log.Fatalf("Error Remove in ProcBaseClnt.getRetStat: %v", err)
		}
	}

	return string(b)
}
