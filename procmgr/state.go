package procmgr

import (
	"sync"

	"sigmaos/proc"
	sp "sigmaos/sigmap"
)

type ProcState struct {
	sync.Mutex
	spawned     map[sp.Tpid]*proc.Proc
	startWaiter map[sp.Tpid]*Waiter
	evictWaiter map[sp.Tpid]*Waiter
	exitWaiter  map[sp.Tpid]*Waiter
	exitStatus  map[sp.Tpid]*ExitStatus
}

func NewProcState() *ProcState {
	return &ProcState{
		spawned:     make(map[sp.Tpid]*proc.Proc),
		startWaiter: make(map[sp.Tpid]*Waiter),
		evictWaiter: make(map[sp.Tpid]*Waiter),
		exitWaiter:  make(map[sp.Tpid]*Waiter),
		exitStatus:  make(map[sp.Tpid]*ExitStatus),
	}
}

func (ps *ProcState) spawn(p *proc.Proc) {
	ps.Lock()
	defer ps.Unlock()

	ps.spawned[p.GetPid()] = p
	ps.startWaiter[p.GetPid()] = newWaiter(&ps.Mutex)
	ps.evictWaiter[p.GetPid()] = newWaiter(&ps.Mutex)
	ps.exitWaiter[p.GetPid()] = newWaiter(&ps.Mutex)
	ps.exitStatus[p.GetPid()] = newExitStatus(p)
}

func (ps *ProcState) started(pid sp.Tpid) {
	ps.Lock()
	defer ps.Unlock()

	// May be called multiple times by procmgr if, for example, the proc crashes
	// shortly after calling Exited().
	if w, ok := ps.startWaiter[pid]; ok {
		w.release()
	}
}

func (ps *ProcState) waitStart(pid sp.Tpid) {
	ps.Lock()
	defer ps.Unlock()

	if w, ok := ps.startWaiter[pid]; ok {
		w.wait()
	}
}

func (ps *ProcState) evict(pid sp.Tpid) {
	ps.Lock()
	defer ps.Unlock()

	if w, ok := ps.evictWaiter[pid]; ok {
		w.release()
	}
}

func (ps *ProcState) waitEvict(pid sp.Tpid) {
	ps.Lock()
	defer ps.Unlock()

	if w, ok := ps.evictWaiter[pid]; ok {
		w.wait()
	}
}

// May be called multiple times by procmgr if, for example, the proc crashes
// shortly after calling Exited().
func (ps *ProcState) exited(pid sp.Tpid, status *proc.Status) {
	ps.Lock()
	defer ps.Unlock()

	if w, ok := ps.exitWaiter[pid]; ok {
		ps.exitStatus[pid].SetStatus(status)
		w.release()
		// Make sure to release start waiter
		ps.startWaiter[pid].release()
		// Make sure to release evict waiters so we don't leak goroutines
		ps.evictWaiter[pid].release()
	}
	// Clean up state
	delete(ps.spawned, pid)
	delete(ps.startWaiter, pid)
	delete(ps.evictWaiter, pid)
	delete(ps.exitWaiter, pid)
}

func (ps *ProcState) waitExit(pid sp.Tpid) *proc.Status {
	ps.Lock()
	defer ps.Unlock()

	// If proc exited already, the waiter may no longer be present.
	if w, ok := ps.exitWaiter[pid]; ok {
		w.wait()
	}
	status, del := ps.exitStatus[pid].GetStatus()
	if del {
		delete(ps.exitStatus, pid)
	}
	return status
}
