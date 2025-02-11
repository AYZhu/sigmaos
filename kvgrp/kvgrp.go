package kvgrp

//
// Starts a group of servers. If nrepl > 0, then the servers form a
// raft group.  If nrepl == 0, then group is either a single
// server. or Clients can wait until the group has configured using
// WaitStarted().
//

import (
	"path"
	"sync"
	"time"

	"sigmaos/crash"
	db "sigmaos/debug"
	"sigmaos/fslib"
	"sigmaos/leaderclnt"
	"sigmaos/perf"
	"sigmaos/proc"
	"sigmaos/replraft"
	"sigmaos/sigmaclnt"
	sp "sigmaos/sigmap"
	"sigmaos/sigmasrv"
)

const (
	GRPCONF  = "-conf"
	GRPELECT = "-elect"
	GRPSEM   = "-sem"
	KVDIR    = sp.NAMED + "kv/"
)

func JobDir(job string) string {
	return path.Join(KVDIR, job)
}

func GrpPath(jobdir string, grp string) string {
	return path.Join(jobdir, grp)
}

func grpConfPath(jobdir, grp string) string {
	return GrpPath(jobdir, grp) + GRPCONF
}

func grpElectPath(jobdir, grp string) string {
	return GrpPath(jobdir, grp) + GRPELECT
}

func grpSemPath(jobdir, grp string) string {
	return GrpPath(jobdir, grp) + GRPSEM
}

type Group struct {
	sync.Mutex
	jobdir string
	grp    string
	ip     string
	myid   int
	*sigmaclnt.SigmaClnt
	ssrv   *sigmasrv.SigmaSrv
	lc     *leaderclnt.LeaderClnt
	isBusy bool
}

func (g *Group) testAndSetBusy() bool {
	g.Lock()
	defer g.Unlock()
	b := g.isBusy
	g.isBusy = true
	return b
}

func (g *Group) clearBusy() {
	g.Lock()
	defer g.Unlock()
	g.isBusy = false
}

func (g *Group) AcquireLeadership() {
	db.DPrintf(db.KVGRP, "%v/%v Try acquire leadership", g.grp, g.myid)
	if err := g.lc.LeadAndFence(nil, []string{g.jobdir}); err != nil {
		db.DFatalf("LeadAndFence err %v", err)
	}
	db.DPrintf(db.KVGRP, "%v/%v Acquired leadership", g.grp, g.myid)
}

func (g *Group) ReleaseLeadership() {
	if err := g.lc.ReleaseLeadership(); err != nil {
		db.DFatalf("release leadership: %v", err)
	}
	db.DPrintf(db.KVGRP, "%v/%v Released leadership", g.grp, g.myid)
}

// For clients to wait unil a group is ready to serve
func WaitStarted(fsl *fslib.FsLib, jobdir, grp string) (*GroupConfig, error) {
	db.DPrintf(db.KVGRP, "WaitStarted: Wait for %v\n", GrpPath(jobdir, grp))
	if _, err := fsl.GetFileWatch(GrpPath(jobdir, grp)); err != nil {
		db.DPrintf(db.KVGRP, "WaitStarted: GetFileWatch %s err %v\n", GrpPath(jobdir, grp), err)
		return nil, err
	}
	cfg := &GroupConfig{}
	if err := fsl.GetFileJson(grpConfPath(jobdir, grp), cfg); err != nil {
		db.DPrintf(db.KVGRP, "WaitStarted: GetFileJson %s err %v\n", grpConfPath(jobdir, grp), err)
		return nil, err
	}
	return cfg, nil
}

func (g *Group) writeSymlink(sigmaAddrs []sp.Taddrs) {
	srvAddrs := make(sp.Taddrs, 0)
	for _, as := range sigmaAddrs {
		addrs := sp.Taddrs{}
		for _, a := range as {
			addrs = append(addrs, a)
		}
		if len(addrs) > 0 {
			srvAddrs = append(srvAddrs, addrs...)
		}
	}
	mnt := sp.MkMountService(srvAddrs)
	db.DPrintf(db.KVGRP, "Advertise %v at %v", mnt, GrpPath(g.jobdir, g.grp))
	if err := g.MkMountSymlink(GrpPath(g.jobdir, g.grp), mnt, g.lc.Lease()); err != nil {
		db.DFatalf("couldn't read replica addrs %v err %v", g.grp, err)
	}
}

func RunMember(job, grp string, public bool, myid, nrepl int) {
	g := &Group{myid: myid, grp: grp, isBusy: true}

	sc, err := sigmaclnt.MkSigmaClnt(sp.Tuname("kv-" + proc.GetPid().String()))
	if err != nil {
		db.DFatalf("MkSigmaClnt %v\n", err)
	}
	g.SigmaClnt = sc
	g.jobdir = JobDir(job)

	g.lc, err = leaderclnt.MakeLeaderClnt(sc.FsLib, grpElectPath(g.jobdir, grp), 0777)
	if err != nil {
		db.DFatalf("MakeLeaderClnt %v\n", err)
	}

	db.DPrintf(db.KVGRP, "Starting replica %d with replication level %v", g.myid, nrepl)

	g.Started()

	ch := make(chan struct{})
	go g.waitExit(ch)

	g.AcquireLeadership()

	cfg := g.readCreateCfg(g.myid, nrepl)

	var raftCfg *replraft.RaftConfig
	if nrepl > 0 {
		cfg, raftCfg = g.makeRaftCfg(cfg, g.myid, nrepl)
	}

	db.DPrintf(db.KVGRP, "Grp config: %v config: %v raftCfg %v", g.myid, cfg, raftCfg)

	cfg, err = g.startServer(cfg, raftCfg)
	if err != nil {
		db.DFatalf("startServer %v\n", err)
	}

	g.writeSymlink(cfg.SigmaAddrs)

	g.ReleaseLeadership()

	crash.Crasher(g.FsLib)
	crash.Partitioner(g.ssrv.SessSrv)
	crash.NetFailer(g.ssrv.SessSrv)

	// Record performance.
	p, err := perf.MakePerf(perf.GROUP)
	if err != nil {
		db.DFatalf("MakePerf err %v\n", err)
	}
	defer p.Done()

	// g.srv.MonitorCPU(nil)

	db.DPrintf(db.KVGRP, "%v/%v: wait ch\n", g.grp, g.myid)

	<-ch

	db.DPrintf(db.KVGRP, "%v/%v: pid %v done\n", g.grp, g.myid, proc.GetPid())

	g.ssrv.SrvExit(proc.MakeStatus(proc.StatusEvicted))
}

// XXX move to procclnt?
func (g *Group) waitExit(ch chan struct{}) {
	for {
		err := g.WaitEvict(proc.GetPid())
		if err != nil {
			db.DPrintf(db.KVGRP, "WaitEvict err %v", err)
			time.Sleep(time.Second)
			continue
		}
		db.DPrintf(db.KVGRP, "waitExit: %v evicted\n", proc.GetPid())
		ch <- struct{}{}
	}
}
