package benchmarks_test

import (
	"io"
	"os"
	"path"
	"time"

	"github.com/stretchr/testify/assert"

	db "sigmaos/debug"
	"sigmaos/fslib"
	"sigmaos/linuxsched"
	"sigmaos/machine"
	"sigmaos/proc"
	"sigmaos/procclnt"
	"sigmaos/procdclnt"
	"sigmaos/rand"
	"sigmaos/realm"
	"sigmaos/semclnt"
	np "sigmaos/sigmap"
	"sigmaos/test"
)

//
// A set of helper functions that we use across our benchmarks.
//

// ========== Proc Helpers ==========

func makeNProcs(n int, prog string, args []string, env []string, ncore proc.Tcore) ([]*proc.Proc, []interface{}) {
	ps := make([]*proc.Proc, 0, n)
	is := make([]interface{}, 0, n)
	for i := 0; i < n; i++ {
		// Note sleep is much shorter, and since we're running "native" the lambda won't actually call Started or Exited for us.
		p := proc.MakeProc(prog, args)
		p.Env = append(p.Env, env...)
		p.SetNcore(ncore)
		ps = append(ps, p)
		is = append(is, p)
	}
	return ps, is
}

func spawnBurstProcs(ts *test.Tstate, ps []*proc.Proc) {
	db.DPrintf("TEST", "Burst-spawning %v procs in chunks of size %v", len(ps), len(ps)/MAX_PARALLEL)
	_, errs := ts.SpawnBurstParallel(ps, len(ps)/MAX_PARALLEL)
	assert.Equal(ts.T, len(errs), 0, "Errors SpawnBurst: %v", errs)
}

func spawnBurstProcs2(ts *test.Tstate, pclnt *procclnt.ProcClnt, ps []*proc.Proc) {
	db.DPrintf("TEST", "Burst-spawning2 %v procs in chunks of size %v", len(ps), len(ps)/MAX_PARALLEL)
	_, errs := pclnt.SpawnBurstParallel(ps, len(ps)/MAX_PARALLEL)
	assert.Equal(ts.T, len(errs), 0, "Errors SpawnBurst: %v", errs)
}

func waitStartProcs(ts *test.Tstate, ps []*proc.Proc) {
	for _, p := range ps {
		err := ts.WaitStart(p.Pid)
		assert.Nil(ts.T, err, "WaitStart: %v", err)
	}
	db.DPrintf("TEST", "%v burst-spawned procs have all started", len(ps))
}

func waitExitProcs(ts *test.Tstate, ps []*proc.Proc) {
	for _, p := range ps {
		status, err := ts.WaitExit(p.Pid)
		assert.Nil(ts.T, err, "WaitStart: %v", err)
		assert.True(ts.T, status.IsStatusOK(), "Bad status: %v", status)
	}
	db.DPrintf("TEST", "%v burst-spawned procs have all started", len(ps))
}

func evictProcs(ts *test.Tstate, ps []*proc.Proc) {
	for _, p := range ps {
		err := ts.Evict(p.Pid)
		assert.Nil(ts.T, err, "Evict: %v", err)
		status, err := ts.WaitExit(p.Pid)
		assert.True(ts.T, status.IsStatusEvicted(), "Bad status evict: %v", status)
	}
}

// ========== Realm Helpers ==========

func countNClusterCores(ts *test.Tstate) {
	// If realms are turned on, find aggregate number of cores across all
	// machines.
	if ts.RunningInRealm() {
		N_CLUSTER_CORES = 0
		db.DPrintf("TEST", "Running with realms")
		fsl := fslib.MakeFsLib("test")
		_, err := fsl.ProcessDir(machine.MACHINES, func(st *np.Stat) (bool, error) {
			cfg := machine.MakeEmptyConfig()
			err := fsl.GetFileJson(path.Join(machine.MACHINES, st.Name, machine.CONFIG), cfg)
			if err != nil {
				return true, err
			}
			N_CLUSTER_CORES += int(cfg.Cores.Size())
			return false, nil
		})
		assert.Nil(ts.T, err, "Error counting sigma cores: %v", err)
	} else {
		db.DPrintf("TEST", "Running without realms")
		N_CLUSTER_CORES = int(linuxsched.NCores)
	}
	db.DPrintf("TEST", "Aggregate number of cores in the cluster: %v", N_CLUSTER_CORES)
}

// Potentially pregrow a realm to encompass all cluster resources.
func maybePregrowRealm(ts *test.Tstate) {
	if PREGROW_REALM {
		// Make sure we've counted the number of cores in the cluster.
		countNClusterCores(ts)
		rclnt := realm.MakeRealmClntFsl(fslib.MakeFsLib("test-rclnt"), ts.ProcClnt)
		// While we are missing cores, try to grow.
		for realm.GetRealmConfig(rclnt.FsLib, ts.RealmId()).NCores != proc.Tcore(N_CLUSTER_CORES) {
			rclnt.GrowRealm(ts.RealmId())
		}
		// Sleep for a bit, so procclnts will take note of the change.
		time.Sleep(2 * np.Conf.Realm.RESIZE_INTERVAL)
		pdc := procdclnt.MakeProcdClnt(ts.FsLib, ts.RealmId())
		n, _, err := pdc.Nprocd()
		assert.Nil(ts.T, err, "Err %v", err)
		db.DPrintf("TEST", "Pre-grew realm, now running with %v procds", n)
	}
}

// ========== Dir Helpers ==========

func makeOutDir(ts *test.Tstate) {
	err := ts.MkDir(OUT_DIR, 0777)
	assert.Nil(ts.T, err, "Couldn't make out dir: %v", err)
}

func rmOutDir(ts *test.Tstate) {
	err := ts.RmDir(OUT_DIR)
	assert.Nil(ts.T, err, "Couldn't rm out dir: %v", err)
}

// ========== Semaphore Helpers ==========

func makeNSemaphores(ts *test.Tstate, n int) ([]*semclnt.SemClnt, []interface{}) {
	ss := make([]*semclnt.SemClnt, 0, n)
	is := make([]interface{}, 0, n)
	for i := 0; i < n; i++ {
		spath := path.Join(OUT_DIR, rand.String(16))
		s := semclnt.MakeSemClnt(ts.FsLib, spath)
		ss = append(ss, s)
		is = append(is, s)
	}
	return ss, is
}

// ========== MR Helpers ========

func makeNMRJobs(ts *test.Tstate, n int, app string) ([]*MRJobInstance, []interface{}) {
	ms := make([]*MRJobInstance, 0, n)
	is := make([]interface{}, 0, n)
	for i := 0; i < n; i++ {
		i := MakeMRJobInstance(ts, app, app+"-mr-"+rand.String(16)+"-"+ts.RealmId())
		ms = append(ms, i)
		is = append(is, i)
	}
	return ms, is
}

// ========== KV Helpers ========

func parseDurations(ts *test.Tstate, ss []string) []time.Duration {
	ds := make([]time.Duration, 0, len(ss))
	for _, s := range ss {
		d, err := time.ParseDuration(s)
		assert.Nil(ts.T, err, "Error parse duration: %v", err)
		ds = append(ds, d)
	}
	return ds
}

func makeNKVJobs(ts *test.Tstate, n, nkvd, kvdrepl int, nclerks []int, phases []time.Duration, ckdur string, kvdncore, ckncore proc.Tcore, auto string, redisaddr string) ([]*KVJobInstance, []interface{}) {
	// If we're running with unbounded clerks...
	if len(phases) > 0 {
		assert.Equal(ts.T, len(nclerks), len(phases), "Phase and clerk lengths don't match: %v != %v", len(phases), len(nclerks))
	}
	js := make([]*KVJobInstance, 0, n)
	is := make([]interface{}, 0, n)
	for i := 0; i < n; i++ {
		ji := MakeKVJobInstance(ts, nkvd, kvdrepl, nclerks, phases, ckdur, kvdncore, ckncore, auto, redisaddr)
		js = append(js, ji)
		is = append(is, ji)
	}
	return js, is
}

// ========== Www Helpers ========

func makeWwwJobs(ts *test.Tstate, sigmaos bool, n int, wwwncore proc.Tcore, reqtype string, ntrials, nclnt, nreq int, delay time.Duration) ([]*WwwJobInstance, []interface{}) {
	ws := make([]*WwwJobInstance, 0, n)
	is := make([]interface{}, 0, n)
	for i := 0; i < n; i++ {
		i := MakeWwwJob(ts, sigmaos, wwwncore, reqtype, ntrials, nclnt, nreq, delay)
		ws = append(ws, i)
		is = append(is, i)
	}
	return ws, is
}

// ========== Hotel Helpers ========

func makeHotelJobs(ts *test.Tstate, sigmaos bool, dur string, maxrps string, fn hotelFn) ([]*HotelJobInstance, []interface{}) {
	// n is ntrials, which is always 1.
	n := 1
	ws := make([]*HotelJobInstance, 0, n)
	is := make([]interface{}, 0, n)
	for i := 0; i < n; i++ {
		i := MakeHotelJob(ts, sigmaos, dur, maxrps, fn)
		ws = append(ws, i)
		is = append(is, i)
	}
	return ws, is
}

// ========== Download Results Helpers ==========

func downloadS3Results(ts *test.Tstate, src string, dst string) {
	// Make the destination directory.
	os.MkdirAll(dst, 0777)
	_, err := ts.ProcessDir(src, func(st *np.Stat) (bool, error) {
		rdr, err := ts.OpenReader(path.Join(src, st.Name))
		defer rdr.Close()
		assert.Nil(ts.T, err, "Error open reader %v", err)
		b, err := io.ReadAll(rdr)
		assert.Nil(ts.T, err, "Error read all %v", err)
		err = os.WriteFile(path.Join(dst, st.Name), b, 0777)
		assert.Nil(ts.T, err, "Error write file %v", err)
		return false, nil
	})
	assert.Nil(ts.T, err, "Error process dir %v", err)
}
