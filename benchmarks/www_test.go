package benchmarks_test

import (
	"path"
	"time"

	"github.com/stretchr/testify/assert"

	db "sigmaos/debug"
	"sigmaos/proc"
	"sigmaos/rand"
	"sigmaos/semclnt"
	sp "sigmaos/sigmap"
	"sigmaos/test"
	"sigmaos/www"
)

const (
	K8S_PORT = ":32585"
)

type WwwJobInstance struct {
	sigmaos    bool
	k8ssrvaddr string
	wwwmcpu    proc.Tmcpu // Number of exclusive cores allocated to each wwwd
	job        string
	ntrials    int
	nclnt      int
	nreq       int
	delay      time.Duration
	ready      chan bool
	sem        *semclnt.SemClnt
	sempath    string
	cpids      []proc.Tpid
	pid        proc.Tpid
	*test.RealmTstate
}

func MakeWwwJob(ts *test.RealmTstate, sigmaos bool, wwwmcpu proc.Tmcpu, reqtype string, ntrials, nclnt, nreq int, delay time.Duration) *WwwJobInstance {
	ji := &WwwJobInstance{}
	ji.sigmaos = sigmaos
	ji.job = rand.String(16)
	ji.ntrials = ntrials
	ji.nclnt = nclnt
	ji.nreq = nreq
	ji.delay = delay
	ji.ready = make(chan bool)
	ji.RealmTstate = ts

	www.InitWwwFs(ts.FsLib, ji.job)

	if !sigmaos {
		db.DFatalf("Error: Get actual machine IP for k8s")
		//		ip, err := fidclnt.LocalIP()
		//		assert.Nil(ji.Ts.T, err, "Error LocalIP: %v", err)
		ip := ""
		ji.k8ssrvaddr = ip + K8S_PORT
	}

	ji.sempath = path.Join(www.JobDir(ji.job), "kvclerk-sem")
	ji.sem = semclnt.MakeSemClnt(ts.FsLib, ji.sempath)
	err := ji.sem.Init(0)
	assert.Nil(ji.Ts.T, err, "Sem init: %v", err)
	assert.Equal(ji.Ts.T, reqtype, "compute")
	return ji
}

func (ji *WwwJobInstance) RunClient(j int, ch chan time.Duration) {
	var clnt *www.WWWClnt
	if ji.sigmaos {
		clnt = www.MakeWWWClnt(ji.FsLib, ji.job)
	} else {
		clnt = www.MakeWWWClntAddr(sp.MkTaddrs([]string{ji.k8ssrvaddr}))
	}
	var latency time.Duration
	for i := 0; i < ji.nreq; i++ {
		slp := ji.delay * 2 * time.Duration(rand.Uint64()%100) / 100
		time.Sleep(slp)
		start := time.Now()
		err := clnt.MatMul(MAT_SIZE)
		assert.Nil(ji.Ts.T, err, "Error matmul %v", err)
		latency += time.Since(start)
	}
	ch <- latency
}

func (ji *WwwJobInstance) StartWwwJob() {
	if ji.sigmaos {
		a := proc.MakeProc("wwwd", []string{ji.job, ""})
		err := ji.Spawn(a)
		assert.Nil(ji.Ts.T, err, "Spawn")
		err = ji.WaitStart(a.GetPid())
		ji.pid = a.GetPid()
		assert.Equal(ji.Ts.T, nil, err)
	}
	db.DPrintf(db.ALWAYS, "StartWwwJob ntrial %v nclnt %v nreq/clnt %v avgdelay %v", ji.ntrials, ji.nclnt, ji.nreq, ji.delay)
	for i := 1; i <= ji.nclnt; i++ {
		for j := 0; j < ji.ntrials; j++ {
			ch := make(chan time.Duration)
			for c := 0; c < i; c++ {
				go ji.RunClient(c, ch)
			}
			var totalLatency time.Duration
			for c := 0; c < i; c++ {
				totalLatency += <-ch
			}
			d := totalLatency.Milliseconds()
			db.DPrintf(db.ALWAYS, "trial %v nclnt %d avg latency %vms", j, i, float64(d)/(float64(ji.nreq)*float64(i)))
			sts, err := ji.GetDir("name/procd/ws/runq-lc")
			if err != nil {
				db.DFatalf("Error getdir: %v", err)
			}
			db.DPrintf(db.ALWAYS, "len ws dir: %v", len(sts))
			if len(sts) > 0 {
				for _, st := range sts {
					db.DPrintf(db.ALWAYS, "ws present: %v, mtime %v", st.Name, st.Mtime)
				}
			}
		}
	}
}

func (ji *WwwJobInstance) Wait() {
	if ji.sigmaos {
		clnt := www.MakeWWWClnt(ji.FsLib, ji.job)
		err := clnt.StopServer(ji.ProcClnt, ji.pid)
		assert.Nil(ji.Ts.T, err)
	}
}
