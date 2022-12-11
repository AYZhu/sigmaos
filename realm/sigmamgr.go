package realm

import (
	"path"
	"sync"
	"sync/atomic"
	"time"

	"sigmaos/config"
	db "sigmaos/debug"
	"sigmaos/electclnt"
	"sigmaos/fslib"
	"sigmaos/machine"
	"sigmaos/memfssrv"
	"sigmaos/proc"
	"sigmaos/procclnt"
	"sigmaos/protdevclnt"
	"sigmaos/realm/proto"
	"sigmaos/resource"
	np "sigmaos/sigmap"
	"sigmaos/stats"
)

type SigmaResourceMgr struct {
	sync.Mutex
	freeCoreGroups int64
	realmCreate    chan string
	realmDestroy   chan string
	realmmgrs      map[string]proc.Tpid
	realmLocks     map[string]*electclnt.ElectClnt
	rclnts         map[string]*protdevclnt.ProtDevClnt
	*procclnt.ProcClnt
	*config.ConfigClnt
	*fslib.FsLib
	*memfssrv.MemFs
}

func MakeSigmaResourceMgr() *SigmaResourceMgr {
	m := &SigmaResourceMgr{}
	m.realmCreate = make(chan string)
	m.realmDestroy = make(chan string)
	var err error
	m.MemFs, m.FsLib, m.ProcClnt, err = memfssrv.MakeMemFs(np.SIGMAMGR, "sigmamgr")
	if err != nil {
		db.DFatalf("Error MakeMemFs: %v", err)
	}
	// Mount the KPIDS dir.
	if err := procclnt.MountPids(m.FsLib, fslib.Named()); err != nil {
		db.DFatalf("Error mountpids: %v", err)
	}
	m.ConfigClnt = config.MakeConfigClnt(m.FsLib)
	m.initFS()
	resource.MakeCtlFile(m.receiveResourceGrant, m.handleResourceRequest, m.Root(), np.RESOURCE_CTL)
	m.realmLocks = make(map[string]*electclnt.ElectClnt)
	m.rclnts = make(map[string]*protdevclnt.ProtDevClnt)
	m.realmmgrs = make(map[string]proc.Tpid)

	return m
}

// Make the initial realm dirs, and remove the unneeded union dirs.
func (m *SigmaResourceMgr) initFS() {
	dirs := []string{
		REALM_CONFIG,
		NODED_CONFIG,
		machine.MACHINES,
		REALM_NAMEDS,
		REALM_FENCES,
		REALM_MGRS,
	}
	for _, d := range dirs {
		if err := m.MkDir(d, 0777); err != nil {
			db.DFatalf("Error Mkdir %v in SigmaResourceMgr.initFs: %v", d, err)
		}
	}
}

func (m *SigmaResourceMgr) receiveResourceGrant(msg *resource.ResourceMsg) {
	switch msg.ResourceType {
	case resource.Trealm:
		m.destroyRealm(msg.Name)
	case resource.Tnode:
		db.DPrintf("SIGMAMGR", "free noded %v", msg.Name)
	case resource.Tcore:
		m.freeCores(1)
		db.DPrintf("SIGMAMGR", "free cores %v", msg.Name)
	default:
		db.DFatalf("Unexpected resource type: %v", msg.ResourceType)
	}
}

// Handle a resource request.
func (m *SigmaResourceMgr) handleResourceRequest(msg *resource.ResourceMsg) {
	switch msg.ResourceType {
	case resource.Trealm:
		m.createRealm(msg.Name)
	case resource.Tcore:
		m.Lock()
		defer m.Unlock()

		realmId := msg.Name
		// If realm still exists, try to grow it.
		if _, ok := m.realmLocks[realmId]; ok {
			m.growRealmL(realmId, msg.Amount)
		}
	default:
		db.DFatalf("Unexpected resource type: %v", msg.ResourceType)
	}
}

// TODO: should probably release lock in this loop.
func (m *SigmaResourceMgr) tryGetFreeCores(nRetries int) bool {
	for i := 0; i < nRetries; i++ {
		if atomic.LoadInt64(&m.freeCoreGroups) > 0 {
			return true
		}
		db.DPrintf("SIGMAMGR", "Tried to get cores, but none free.")
		// TODO: parametrize?
		time.Sleep(10 * time.Millisecond)
	}
	db.DPrintf("SIGMAMGR", "Failed to find any free cores.")
	return false
}

func (m *SigmaResourceMgr) allocCores(realmId string, i int64) {
	atomic.AddInt64(&m.freeCoreGroups, -1*i)
	res := &proto.RealmMgrResponse{}
	req := &proto.RealmMgrRequest{
		Ncores: i,
	}
	err := m.rclnts[realmId].RPC("RealmMgr.GrantCores", req, res)
	if err != nil || !res.OK {
		db.DFatalf("Error RPC: %v %v", err, res.OK)
	}
}

func (m *SigmaResourceMgr) freeCores(i int64) {
	atomic.AddInt64(&m.freeCoreGroups, i)
}

// Tries to add a Noded to a realm. Will first try and pull from the list of
// free Nodeds, and if none is available, it will try to make one free, and
// then retry. Caller holds lock.
func (m *SigmaResourceMgr) growRealmL(realmId string, qlen int) bool {
	// See if any cores are available.
	if m.tryGetFreeCores(1) {
		// Try to alloc qlen cores, or as many as are currently free otherwise.
		nallocd := int64(qlen)
		if nallocd == 0 {
			nallocd = 1
		}
		nfree := atomic.LoadInt64(&m.freeCoreGroups)
		if nfree < nallocd {
			nallocd = nfree
		}
		db.DPrintf("SIGMAMGR", "Allocate %v free cores", nallocd)
		// Allocate cores to this realm.
		if nallocd > 0 {
			m.allocCores(realmId, nallocd)
			return true
		}
	}
	// No cores were available, so try to find a realm with spare resources.
	opRealmId, ok := m.findOverProvisionedRealm(realmId)
	if !ok {
		db.DPrintf("SIGMAMGR", "No overprovisioned realms available")
		return false
	}
	// Ask the over-provisioned realm to give up some cores.
	m.requestCores(opRealmId)
	// Wait for the over-provisioned realm to cede its cores.
	if m.tryGetFreeCores(100) {
		// Allocate core to this realm.
		m.allocCores(realmId, 1)
		return true
	}
	return false
}

// Ascertain whether or not a noded is overprovisioned.
//
// XXX Eventually, we'll want to find overprovisioned realms according to
// more nuanced metrics, e.g. how many Nodeds are running procs that hold
// state, etc.
func nodedOverprovisioned(fsl *fslib.FsLib, cc *config.ConfigClnt, realmId string, nodedId string, debug string) bool {
	ndCfg := MakeNodedConfig()
	cc.ReadConfig(NodedConfPath(nodedId), ndCfg)
	db.DPrintf(debug, "Check if noded %v realm %v is overprovisioned", nodedId, realmId)
	s := &stats.StatInfo{}
	err := fsl.GetFileJson(path.Join(RealmPath(realmId), np.PROCDREL, ndCfg.ProcdIp, np.STATSD), s)
	// Only overprovisioned if hasn't shut down/crashed.
	if err != nil {
		db.DPrintf(debug+"_ERR", "Error ReadFileJson in SigmaResourceMgr.getRealmProcdStats: %v", err)
		return false
	}
	// Count the total number of cores assigned to this noded.
	totalCores := 0.0
	for _, cores := range ndCfg.Cores {
		totalCores += float64(cores.Size())
	}
	nLCCoresUsed := s.CustomUtil / 100.0
	// Count how many cores we would revoke.
	coresToRevoke := float64(ndCfg.Cores[len(ndCfg.Cores)-1].Size())
	// If we don't have >= 1 core group to spare for LC procs, we aren't
	// overprovisioned
	if totalCores-coresToRevoke < nLCCoresUsed {
		db.DPrintf(debug, "Noded is using LC cores well, not overprovisioned: %v - %v >= %v", totalCores, coresToRevoke, nLCCoresUsed)
		return false
	}
	db.DPrintf(debug, "Noded %v has %v cores remaining.", nodedId, len(ndCfg.Cores))
	// Don't evict this noded if it is running any LC procs.
	if len(ndCfg.Cores) == 1 {
		qs := []string{np.PROCD_RUNQ_LC}
		for _, q := range qs {
			queued, err := fsl.GetDir(path.Join(RealmPath(realmId), np.PROCDREL, ndCfg.ProcdIp, q))
			if err != nil {
				db.DPrintf(debug+"_ERR", "Couldn't get queue dir %v: %v", q, err)
				return false
			}
			// If there are LC procs queued, don't shrink.
			if len(queued) > 0 {
				db.DPrintf(debug, "Can't evict noded, had %v queued LC procs", len(queued))
				return false
			}
		}
		runningProcs, err := fsl.GetDir(path.Join(RealmPath(realmId), np.PROCDREL, ndCfg.ProcdIp, np.PROCD_RUNNING))
		if err != nil {
			db.DPrintf(debug+"_ERR", "Couldn't get procs running dir: %v", err)
			return false
		}
		// If this is the last core group for this noded, and its utilization is over
		// a certain threshold (and it is running procs), don't evict.
		if s.Util >= np.Conf.Realm.SHRINK_CPU_UTIL_THRESHOLD && len(runningProcs) > 0 {
			db.DPrintf(debug, "Can't evict noded, util: %v runningProcs: %v", s.Util, len(runningProcs))
			return false
		}
		for _, st := range runningProcs {
			p := proc.MakeEmptyProc()
			err := fsl.GetFileJson(path.Join(RealmPath(realmId), np.PROCDREL, ndCfg.ProcdIp, np.PROCD_RUNNING, st.Name), p)
			if err != nil {
				continue
			}
			// If this is a LC proc, return false.
			if p.Type == proc.T_LC {
				db.DPrintf(debug, "Can't evict noded, running LC proc")
				return false
			} else {
				db.DPrintf(debug, "Noded %v's proc %v, is not LC", nodedId, p)
			}
		}
		db.DPrintf(debug, "Evicting noded %v realm %v", nodedId, realmId)
	}
	return true
}

// Find an over-provisioned realm (a realm with resources to spare). Returns
// true if an overprovisioned realm was found, false otherwise.
func (m *SigmaResourceMgr) findOverProvisionedRealm(ignoreRealm string) (string, bool) {
	opRealmId := ""
	ok := false
	m.ProcessDir(REALM_CONFIG, func(st *np.Stat) (bool, error) {
		realmId := st.Name

		// Don't steal a noded from the requesting realm.
		if realmId == ignoreRealm {
			return false, nil
		}

		lock, exists := m.realmLocks[realmId]
		// If the realm we are looking at has been deleted, move on.
		if !exists {
			return false, nil
		}

		lockRealm(lock, realmId)
		defer unlockRealm(lock, realmId)

		rCfg := &RealmConfig{}
		m.ReadConfig(RealmConfPath(realmId), rCfg)

		// See if any nodeds have cores to spare.
		overprovisioned := false
		for _, nd := range rCfg.NodedsAssigned {
			if nodedOverprovisioned(m.FsLib, m.ConfigClnt, realmId, nd, "SIGMAMGR") {
				overprovisioned = true
				break
			}
		}
		// If there are more than the minimum number of required Nodeds available...
		if len(rCfg.NodedsAssigned) > nReplicas() && overprovisioned {
			opRealmId = realmId
			ok = true
			return true, nil
		}
		return false, nil
	})
	return opRealmId, ok
}

// Create a realm.
func (m *SigmaResourceMgr) createRealm(realmId string) {
	m.Lock()
	defer m.Unlock()

	// Make sure we haven't created this realm before.
	if _, ok := m.realmLocks[realmId]; ok {
		db.DFatalf("tried to create realm twice %v", realmId)
	}
	m.realmLocks[realmId] = electclnt.MakeElectClnt(m.FsLib, realmFencePath(realmId), 0777)

	lockRealm(m.realmLocks[realmId], realmId)

	cfg := &RealmConfig{}
	cfg.Rid = realmId

	// Make the realm config file.
	m.WriteConfig(RealmConfPath(realmId), cfg)

	unlockRealm(m.realmLocks[realmId], realmId)

	// Start this realm's realmmgr.
	m.startRealmMgr(realmId)
}

// Request a Noded from realm realmId.
func (m *SigmaResourceMgr) requestCores(realmId string) {
	db.DPrintf("SIGMAMGR", "Sigmamgr requesting cores from %v", realmId)
	res := &proto.RealmMgrResponse{}
	req := &proto.RealmMgrRequest{
		Ncores: 1,
	}
	err := m.rclnts[realmId].RPC("RealmMgr.RevokeCores", req, res)
	if err != nil || !res.OK {
		db.DFatalf("Error RPC: %v %v", err, res.OK)
	}
	db.DPrintf("SIGMAMGR", "Sigmamgr done requesting cores from %v", realmId)
}

// Destroy a realm.
func (m *SigmaResourceMgr) destroyRealm(realmId string) {
	m.Lock()
	defer m.Unlock()

	db.DPrintf("SIGMAMGR", "Destroy realm %v", realmId)

	lockRealm(m.realmLocks[realmId], realmId)

	// Update the realm config to note that the realm is being shut down.
	cfg := &RealmConfig{}
	m.ReadConfig(RealmConfPath(realmId), cfg)
	cfg.Shutdown = true
	m.WriteConfig(RealmConfPath(realmId), cfg)

	unlockRealm(m.realmLocks[realmId], realmId)
	delete(m.realmLocks, realmId)

	res := &proto.RealmMgrResponse{}
	req := &proto.RealmMgrRequest{
		AllCores: true,
	}
	err := m.rclnts[realmId].RPC("RealmMgr.ShutdownRealm", req, res)
	if err != nil || !res.OK {
		db.DFatalf("Error RPC: %v %v", err, res.OK)
	}

	m.evictRealmMgr(realmId)
	db.DPrintf("SIGMAMGR", "Done destroying realm %v", realmId)
}

func (m *SigmaResourceMgr) startRealmMgr(realmId string) {
	pid := proc.Tpid("realmmgr-" + proc.GenPid().String())
	p := proc.MakeProcPid(pid, "realm/realmmgr", []string{realmId})
	if _, err := m.SpawnKernelProc(p, fslib.Named(), "", false); err != nil {
		db.DFatalf("Error spawn realmmgr %v", err)
	}
	if err := m.WaitStart(p.Pid); err != nil {
		db.DFatalf("Error WaitStart realmmgr %v", err)
	}
	db.DPrintf("SIGMAMGR", "Sigmamgr started realmmgr %v in realm %v", pid.String(), realmId)
	m.realmmgrs[realmId] = pid
	var err error
	m.rclnts[realmId], err = protdevclnt.MkProtDevClnt(m.FsLib, realmMgrPath(realmId))
	if err != nil {
		db.DFatalf("Error MkProtDevClnt: %v", err)
	}
}

func (m *SigmaResourceMgr) evictRealmMgr(realmId string) {
	pid := m.realmmgrs[realmId]
	db.DPrintf("SIGMAMGR", "Sigmamgr evicting realmmgr %v in realm %v", pid.String(), realmId)
	if err := m.Evict(pid); err != nil {
		db.DFatalf("Error evict realmmgr %v for realm %v", pid, realmId)
	}
	if status, err := m.WaitExit(pid); err != nil || !status.IsStatusEvicted() {
		db.DFatalf("Error bad status evict realmmgr %v for realm %v: status %v err %v", pid, realmId, status, err)
	}
	delete(m.realmmgrs, realmId)
	delete(m.rclnts, realmId)
}

func (m *SigmaResourceMgr) Work() {
	m.Serve()
	m.Done()
}
