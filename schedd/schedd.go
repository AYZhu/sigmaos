package schedd

import (
	"path"
	"sync"
	"time"

	db "sigmaos/debug"
	"sigmaos/fs"
	"sigmaos/linuxsched"
	"sigmaos/mem"
	"sigmaos/memfssrv"
	"sigmaos/perf"
	"sigmaos/proc"
	"sigmaos/procmgr"
	"sigmaos/schedd/proto"
	"sigmaos/scheddclnt"
	sp "sigmaos/sigmap"
	"sigmaos/sigmasrv"
)

type Schedd struct {
	mu         sync.Mutex
	qsmu       sync.RWMutex
	cond       *sync.Cond
	pmgr       *procmgr.ProcMgr
	scheddclnt *scheddclnt.ScheddClnt
	mcpufree   proc.Tmcpu
	memfree    proc.Tmem
	qs         map[sp.Trealm]*Queue
	kernelId   string
	realms     []sp.Trealm
}

func MakeSchedd(mfs *memfssrv.MemFs, kernelId string, reserveMcpu uint) *Schedd {
	sd := &Schedd{
		pmgr:     procmgr.MakeProcMgr(mfs, kernelId),
		qs:       make(map[sp.Trealm]*Queue),
		realms:   make([]sp.Trealm, 0),
		mcpufree: proc.Tmcpu(1000 * linuxsched.NCores - reserveMcpu),
		memfree:  mem.GetTotalMem(),
		kernelId: kernelId,
	}
	sd.cond = sync.NewCond(&sd.mu)
	sd.scheddclnt = scheddclnt.MakeScheddClnt(mfs.SigmaClnt().FsLib)
	return sd
}

func (sd *Schedd) Spawn(ctx fs.CtxI, req proto.SpawnRequest, res *proto.SpawnResponse) error {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	p := proc.MakeProcFromProto(req.ProcProto)
	p.KernelId = sd.kernelId
	db.DPrintf(db.SCHEDD, "[%v] %v Spawned %v", req.Realm, sd.kernelId, p)
	realm := sp.Trealm(req.Realm)
	q, ok := sd.getQueue(realm)
	if !ok {
		q = sd.addRealmQueueL(realm)
	}
	// Enqueue the proc according to its realm
	q.Enqueue(p)
	s := time.Now()
	sd.pmgr.Spawn(p)
	db.DPrintf(db.SPAWN_LAT, "[%v] E2E Procmgr Spawn %v", p.GetPid(), time.Since(s))
	// Signal that a new proc may be runnable.
	sd.cond.Signal()
	return nil
}

// Steal a proc from this schedd.
func (sd *Schedd) StealProc(ctx fs.CtxI, req proto.StealProcRequest, res *proto.StealProcResponse) error {
	q, _ := sd.getQueue(sp.Trealm(req.Realm))
	_, res.OK = q.Steal(proc.Tpid(req.PidStr))

	return nil
}

// Steal a proc from this schedd.
func (sd *Schedd) GetCPUShares(ctx fs.CtxI, req proto.GetCPUSharesRequest, res *proto.GetCPUSharesResponse) error {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	sm := sd.pmgr.GetCPUShares()
	smap := make(map[string]int64, len(sm))
	for r, s := range sm {
		smap[r.String()] = int64(s)
	}
	res.Shares = smap
	return nil
}

// Steal a proc from this schedd.
func (sd *Schedd) GetCPUUtil(ctx fs.CtxI, req proto.GetCPUUtilRequest, res *proto.GetCPUUtilResponse) error {
	res.Util = sd.pmgr.GetCPUUtil(sp.Trealm(req.RealmStr))
	return nil
}

func (sd *Schedd) procDone(p *proc.Proc) error {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	db.DPrintf(db.SCHEDD, "Proc done %v", p)
	sd.freeResourcesL(p)
	// Signal that a new proc may be runnable.
	sd.cond.Signal()
	return nil
}

// Run a proc via the local procd.
func (sd *Schedd) runProc(p *proc.Proc) {
	sd.allocResourcesL(p)
	go func() {
		sd.pmgr.RunProc(p)
		sd.procDone(p)
	}()
}

func (sd *Schedd) schedule() {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	// Priority order in which procs are claimed
	priority := []proc.Ttype{proc.T_LC, proc.T_BE}
	for {
		var ok bool
		// Iterate through the realms round-robin.
		for _, ptype := range priority {
			for r, q := range sd.qs {
				// Try to schedule a proc from realm r.
				ok = ok || sd.tryScheduleRealmL(r, q, ptype)
			}
			// If a proc was successfully scheduled, don't try to schedule a proc
			// from a lower priority class. Instead, rerun the whole scheduling loop.
			if ok {
				break
			}
		}
		// If unable to schedule a proc from any realm, wait.
		if !ok {
			db.DPrintf(db.SCHEDD, "No procs runnable mcpu:%v mem:%v qs:%v", sd.mcpufree, sd.memfree, sd.qs)
			sd.cond.Wait()
		}
	}
}

// Try to schedule a proc from realm r's queue q. Returns true if a proc was
// successfully scheduled.
func (sd *Schedd) tryScheduleRealmL(r sp.Trealm, q *Queue, ptype proc.Ttype) bool {
	for {
		// Try to dequeue a proc, whether it be from a local queue or potentially
		// stolen from a remote queue.
		if p, stolen, ok := q.Dequeue(ptype, sd.mcpufree, sd.memfree); ok {
			// If the proc was stolen...
			if stolen {
				// Try to claim the proc.
				if ok := sd.tryStealProc(r, p); ok {
					// Proc was claimed successfully.
					db.DPrintf(db.SCHEDD, "[%v] stole proc %v", r, p)
				} else {
					// Couldn't claim the proc. Try and steal another.
					continue
				}
			}
			// Claimed a proc, so schedule it.
			db.DPrintf(db.SCHEDD, "[%v] run proc %v", r, p)
			db.DPrintf(db.SPAWN_LAT, "[%v] Queueing latency %v", p.GetPid(), time.Since(p.GetSpawnTime()))
			sd.runProc(p)
			return true
		} else {
			return false
		}
	}
}

func (sd *Schedd) getQueue(realm sp.Trealm) (*Queue, bool) {
	sd.qsmu.RLock()
	defer sd.qsmu.RUnlock()

	q, ok := sd.qs[realm]
	return q, ok
}

// Caller must hold sd.mu to be held.
func (sd *Schedd) addRealmQueueL(realm sp.Trealm) *Queue {
	sd.qsmu.Lock()
	defer sd.qsmu.Unlock()

	q := makeQueue()
	sd.qs[realm] = q
	return q
}

func RunSchedd(kernelId string, reserveMcpu uint) error {
	mfs, err := memfssrv.MakeMemFs(path.Join(sp.SCHEDD, kernelId), sp.SCHEDDREL)
	if err != nil {
		db.DFatalf("Error MakeMemFs: %v", err)
	}
	sd := MakeSchedd(mfs, kernelId, reserveMcpu)
	ssrv, err := sigmasrv.MakeSigmaSrvMemFs(mfs, sd)
	if err != nil {
		db.DFatalf("Error PDS: %v", err)
	}
	setupMemFsSrv(ssrv.MemFs)
	setupFs(ssrv.MemFs, sd)
	// Perf monitoring
	p, err := perf.MakePerf(perf.SCHEDD)
	if err != nil {
		db.DFatalf("Error MakePerf: %v", err)
	}
	defer p.Done()
	go sd.schedule()
	go sd.monitorWSQueue(proc.T_LC)
	go sd.monitorWSQueue(proc.T_BE)
	go sd.offerStealableProcs()
	ssrv.RunServer()
	return nil
}
