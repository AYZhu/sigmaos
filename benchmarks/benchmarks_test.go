package benchmarks_test

import (
	"flag"
	"math/rand"
	"testing"
	"time"

	// XXX Used for matrix tests before.
	//	"fmt"

	"sigmaos/benchmarks"
	db "sigmaos/debug"
	"sigmaos/hotel"
	"sigmaos/linuxsched"
	"sigmaos/proc"
	"sigmaos/test"
)

// Parameters
var N_TRIALS int
var PREGROW_REALM bool
var MR_APP string
var KV_AUTO string
var N_KVD int
var N_CLERK int
var CLERK_DURATION string
var CLERK_NCORE int
var N_CLNT int
var N_CLNT_REQ int
var KVD_NCORE int
var WWWD_NCORE int
var WWWD_REQ_TYPE string
var WWWD_REQ_DELAY time.Duration
var HOTEL_NCORE int
var HOTEL_DUR time.Duration
var HOTEL_MAX_RPS int
var REALM2 string
var REDIS_ADDR string
var N_PROC int
var N_CORE int
var MAT_SIZE int
var CONTENDERS_FRAC float64
var GO_MAX_PROCS int
var MAX_PARALLEL int
var K8S_ADDR string

// XXX REMOVE EVENTUALLY
var AAA int

// Read & set the proc version.
func init() {
	flag.IntVar(&N_TRIALS, "ntrials", 1, "Number of trials.")
	flag.BoolVar(&PREGROW_REALM, "pregrow_realm", false, "Pre-grow realm to include all cluster resources.")
	flag.StringVar(&MR_APP, "mrapp", "mr-wc-wiki1.8G.yml", "Name of mr yaml file.")
	flag.StringVar(&KV_AUTO, "kvauto", "manual", "KV auto-growing/shrinking.")
	flag.IntVar(&N_KVD, "nkvd", 1, "Number of kvds.")
	flag.IntVar(&N_CLERK, "nclerk", 1, "Number of clerks.")
	flag.IntVar(&N_CLNT, "nclnt", 1, "Number of www clients.")
	flag.IntVar(&N_CLNT_REQ, "nclnt_req", 1, "Number of request each www client makes.")
	flag.StringVar(&CLERK_DURATION, "clerk_dur", "90s", "Clerk duration.")
	flag.IntVar(&CLERK_NCORE, "clerk_ncore", 1, "Clerk Ncore")
	flag.IntVar(&KVD_NCORE, "kvd_ncore", 2, "KVD Ncore")
	flag.IntVar(&WWWD_NCORE, "wwwd_ncore", 2, "WWWD Ncore")
	flag.StringVar(&WWWD_REQ_TYPE, "wwwd_req_type", "compute", "WWWD request type [compute, dummy, io].")
	flag.DurationVar(&WWWD_REQ_DELAY, "wwwd_req_delay", 500*time.Millisecond, "Average request delay.")
	flag.IntVar(&HOTEL_NCORE, "hotel_ncore", 1, "Hotel Ncore.")
	flag.DurationVar(&HOTEL_DUR, "hotel_dur", 10*time.Second, "Hotel benchmark load generation duration.")
	flag.IntVar(&HOTEL_MAX_RPS, "hotel_max_rps", 1000, "Max requests/second for hotel bench.")
	flag.StringVar(&K8S_ADDR, "k8saddr", "", "Kubernetes frontend service address (only for hotel benchmarking for the time being).")
	flag.StringVar(&REALM2, "realm2", "test-realm", "Second realm")
	flag.StringVar(&REDIS_ADDR, "redisaddr", "", "Redis server address")
	flag.IntVar(&N_PROC, "nproc", 1, "Number of procs per trial.")
	flag.IntVar(&N_CORE, "ncore", 1, "Generic proc test Ncore")
	flag.IntVar(&MAT_SIZE, "matrixsize", 4000, "Size of matrix.")
	flag.Float64Var(&CONTENDERS_FRAC, "contenders", 4000, "Fraction of cores which should be taken up by contending procs.")
	flag.IntVar(&GO_MAX_PROCS, "gomaxprocs", int(linuxsched.NCores), "Go maxprocs setting for procs to be spawned.")
	flag.IntVar(&MAX_PARALLEL, "max_parallel", 1, "Max amount of parallelism.")
	// XXX Remove after protoyping
	flag.IntVar(&AAA, "aaa", 1, "Num procclnts.")
}

// ========== Common parameters ==========
const (
	OUT_DIR = "name/out_dir"
)

var N_CLUSTER_CORES = 0

// XXX Switch to spin.
//// Length of time required to do a simple matrix multiplication.
//func TestNiceMatMulBaseline(t *testing.T) {
//	ts := test.MakeTstateAll(t)
//	rs := benchmarks.MakeResults(N_TRIALS)
//	_, ps := makeNProcs(N_TRIALS, "user/matmul", []string{fmt.Sprintf("%v", MAT_SIZE)}, []string{fmt.Sprintf("GOMAXPROCS=%v", GO_MAX_PROCS)}, 1)
//	runOps(ts, ps, runProc, rs)
//	printResultSummary(rs)
//	ts.Shutdown()
//}
//
//// Start a bunch of spinning procs to contend with one matmul task, and then
//// see how long the matmul task took.
//func TestNiceMatMulWithSpinners(t *testing.T) {
//	ts := test.MakeTstateAll(t)
//	rs := benchmarks.MakeResults(N_TRIALS)
//	makeOutDir(ts)
//	nContenders := int(float64(linuxsched.NCores) / CONTENDERS_FRAC)
//	// Make some spinning procs to take up nContenders cores.
//	psSpin, _ := makeNProcs(nContenders, "user/spinner", []string{OUT_DIR}, []string{fmt.Sprintf("GOMAXPROCS=%v", 1)}, 0)
//	// Burst-spawn BE procs
//	spawnBurstProcs(ts, psSpin)
//	// Wait for the procs to start
//	waitStartProcs(ts, psSpin)
//	// Make the LC proc.
//	_, ps := makeNProcs(N_TRIALS, "user/matmul", []string{fmt.Sprintf("%v", MAT_SIZE)}, []string{fmt.Sprintf("GOMAXPROCS=%v", GO_MAX_PROCS)}, 1)
//	// Spawn the LC procs
//	runOps(ts, ps, runProc, rs)
//	printResultSummary(rs)
//	evictProcs(ts, psSpin)
//	rmOutDir(ts)
//	ts.Shutdown()
//}
//
//// Invert the nice relationship. Make spinners high-priority, and make matul
//// low priority. This is intended to verify that changing priorities does
//// actually affect application throughput for procs which have their priority
//// lowered, and by how much.
//func TestNiceMatMulWithSpinnersLCNiced(t *testing.T) {
//	ts := test.MakeTstateAll(t)
//	rs := benchmarks.MakeResults(N_TRIALS)
//	makeOutDir(ts)
//	nContenders := int(float64(linuxsched.NCores) / CONTENDERS_FRAC)
//	// Make some spinning procs to take up nContenders cores. (AS LC)
//	psSpin, _ := makeNProcs(nContenders, "user/spinner", []string{OUT_DIR}, []string{fmt.Sprintf("GOMAXPROCS=%v", 1)}, 1)
//	// Burst-spawn spinning procs
//	spawnBurstProcs(ts, psSpin)
//	// Wait for the procs to start
//	waitStartProcs(ts, psSpin)
//	// Make the matmul procs.
//	_, ps := makeNProcs(N_TRIALS, "user/matmul", []string{fmt.Sprintf("%v", MAT_SIZE)}, []string{fmt.Sprintf("GOMAXPROCS=%v", GO_MAX_PROCS)}, 0)
//	// Spawn the matmul procs
//	runOps(ts, ps, runProc, rs)
//	printResultSummary(rs)
//	evictProcs(ts, psSpin)
//	rmOutDir(ts)
//	ts.Shutdown()
//}

// Test how long it takes to init a semaphore.
func TestMicroInitSemaphore(t *testing.T) {
	ts := test.MakeTstateAll(t)
	rs := benchmarks.MakeResults(N_TRIALS, benchmarks.OPS)
	makeOutDir(ts)
	_, is := makeNSemaphores(ts, N_TRIALS)
	runOps(ts, is, initSemaphore, rs)
	printResultSummary(rs)
	rmOutDir(ts)
	ts.Shutdown()
}

// Test how long it takes to up a semaphore.
func TestMicroUpSemaphore(t *testing.T) {
	ts := test.MakeTstateAll(t)
	rs := benchmarks.MakeResults(N_TRIALS, benchmarks.OPS)
	makeOutDir(ts)
	_, is := makeNSemaphores(ts, N_TRIALS)
	// Init semaphores first.
	for _, i := range is {
		initSemaphore(ts, i)
	}
	runOps(ts, is, upSemaphore, rs)
	printResultSummary(rs)
	rmOutDir(ts)
	ts.Shutdown()
}

// Test how long it takes to down a semaphore.
func TestMicroDownSemaphore(t *testing.T) {
	ts := test.MakeTstateAll(t)
	rs := benchmarks.MakeResults(N_TRIALS, benchmarks.OPS)
	makeOutDir(ts)
	_, is := makeNSemaphores(ts, N_TRIALS)
	// Init semaphores first.
	for _, i := range is {
		initSemaphore(ts, i)
		upSemaphore(ts, i)
	}
	runOps(ts, is, downSemaphore, rs)
	printResultSummary(rs)
	rmOutDir(ts)
	ts.Shutdown()
}

// Test how long it takes to Spawn, run, and WaitExit a 5ms proc.
func TestMicroSpawnWaitExit5msSleeper(t *testing.T) {
	ts := test.MakeTstateAll(t)
	rs := benchmarks.MakeResults(N_TRIALS, benchmarks.OPS)
	makeOutDir(ts)
	_, ps := makeNProcs(N_TRIALS, "user/sleeper", []string{"5000us", OUT_DIR}, []string{}, 1)
	runOps(ts, ps, runProc, rs)
	printResultSummary(rs)
	rmOutDir(ts)
	ts.Shutdown()
}

// Test the throughput of spawning procs.
func TestMicroSpawnBurstTpt(t *testing.T) {
	ts := test.MakeTstateAll(t)
	maybePregrowRealm(ts)
	rs := benchmarks.MakeResults(N_TRIALS, benchmarks.OPS)
	db.DPrintf(db.ALWAYS, "SpawnBursting %v procs (ncore=%v) with max parallelism %v", N_PROC, N_CORE, MAX_PARALLEL)
	ps, _ := makeNProcs(N_PROC, "user/sleeper", []string{"0s", ""}, []string{}, proc.Tcore(N_CORE))
	runOps(ts, []interface{}{ps}, spawnBurstWaitStartProcs, rs)
	printResultSummary(rs)
	waitExitProcs(ts, ps)
	ts.Shutdown()
}

func TestAppMR(t *testing.T) {
	ts := test.MakeTstateAll(t)
	countNClusterCores(ts)
	maybePregrowRealm(ts)
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	jobs, apps := makeNMRJobs(ts, 1, MR_APP)
	// XXX Clean this up/hide this somehow.
	go func() {
		for _, j := range jobs {
			// Wait until ready
			<-j.ready
			// Ack to allow the job to proceed.
			j.ready <- true
		}
	}()
	p := monitorCoresAssigned(ts)
	defer p.Done()
	runOps(ts, apps, runMR, rs)
	printResultSummary(rs)
	ts.Shutdown()
}

func runKVTest(t *testing.T, nReplicas int) {
	ts := test.MakeTstateAll(t)
	countNClusterCores(ts)
	maybePregrowRealm(ts)
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	nclerks := []int{N_CLERK}
	db.DPrintf(db.ALWAYS, "Running with %v clerks", N_CLERK)
	jobs, ji := makeNKVJobs(ts, 1, N_KVD, nReplicas, nclerks, nil, CLERK_DURATION, proc.Tcore(KVD_NCORE), proc.Tcore(CLERK_NCORE), KV_AUTO, REDIS_ADDR)
	// XXX Clean this up/hide this somehow.
	go func() {
		for _, j := range jobs {
			// Wait until ready
			<-j.ready
			// Ack to allow the job to proceed.
			j.ready <- true
		}
	}()
	p := monitorCoresAssigned(ts)
	defer p.Done()
	runOps(ts, ji, runKV, rs)
	printResultSummary(rs)
	ts.Shutdown()
}

func TestAppKVUnrepl(t *testing.T) {
	runKVTest(t, 0)
}

func TestAppKVRepl(t *testing.T) {
	runKVTest(t, 3)
}

// Burst a bunch of spinning procs, and see how long it takes for all of them
// to start.
func TestRealmBurst(t *testing.T) {
	ts := test.MakeTstateAll(t)
	countNClusterCores(ts)
	maybePregrowRealm(ts)
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	makeOutDir(ts)
	// Find the total number of cores available for spinners across all machines.
	// We need to get this in order to find out how many spinners to start.
	db.DPrintf(db.ALWAYS, "Bursting %v spinning procs", N_CLUSTER_CORES)
	ps, _ := makeNProcs(N_CLUSTER_CORES, "user/spinner", []string{OUT_DIR}, []string{}, 1)
	p := monitorCoresAssigned(ts)
	defer p.Done()
	runOps(ts, []interface{}{p}, spawnBurstWaitStartProcs, rs)
	printResultSummary(rs)
	evictProcs(ts, ps)
	rmOutDir(ts)
	ts.Shutdown()
}

func TestLambdaBurst(t *testing.T) {
	ts := test.MakeTstateAll(t)
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	makeOutDir(ts)
	// Find the total number of cores available for spinners across all machines.
	// We need to get this in order to find out how many spinners to start.
	N_LAMBDAS := 720
	db.DPrintf(db.ALWAYS, "Invoking %v lambdas", N_LAMBDAS)
	ss, is := makeNSemaphores(ts, N_LAMBDAS)
	// Init semaphores first.
	for _, i := range is {
		initSemaphore(ts, i)
	}
	runOps(ts, []interface{}{ss}, invokeWaitStartLambdas, rs)
	printResultSummary(rs)
	rmOutDir(ts)
	ts.Shutdown()
}

func TestLambdaInvokeWaitStart(t *testing.T) {
	ts := test.MakeTstateAll(t)
	rs := benchmarks.MakeResults(720, benchmarks.E2E)
	makeOutDir(ts)
	// Find the total number of cores available for spinners across all machines.
	// We need to get this in order to find out how many spinners to start.
	N_LAMBDAS := 640
	db.DPrintf(db.ALWAYS, "Invoking %v lambdas", N_LAMBDAS)
	_, is := makeNSemaphores(ts, N_LAMBDAS)
	// Init semaphores first.
	for _, i := range is {
		initSemaphore(ts, i)
	}
	runOps(ts, is, invokeWaitStartOneLambda, rs)
	printResultSummary(rs)
	rmOutDir(ts)
	ts.Shutdown()
}

// Start a realm with a long-running BE mr job. Then, start a realm with an LC
// hotel job. In phases, ramp the hotel job's CPU utilization up and down, and
// watch the realm-level software balance resource requests across realms.
func TestRealmBalanceMRHotel(t *testing.T) {
	done := make(chan bool)
	// Find the total number of cores available for spinners across all machines.
	ts := test.MakeTstateAll(t)
	countNClusterCores(ts)
	// Structures for mr
	ts1 := test.MakeTstateRealm(t, ts.RealmId())
	rs1 := benchmarks.MakeResults(1, benchmarks.E2E)
	// Structure for kv
	ts2 := test.MakeTstateRealm(t, REALM2)
	rs2 := benchmarks.MakeResults(1, benchmarks.E2E)
	// Prep MR job
	mrjobs, mrapps := makeNMRJobs(ts1, 1, MR_APP)
	// Prep Hotel job
	hotelJobs, ji := makeHotelJobs(ts2, true, proc.Tcore(HOTEL_NCORE), HOTEL_DUR, HOTEL_MAX_RPS, func(wc *hotel.WebClnt, r *rand.Rand) {
		//		hotel.RunDSB(ts2.T, 1, wc, r)
		hotel.RandSearchReq(wc, r)
	})
	p1 := monitorCoresAssigned(ts1)
	defer p1.Done()
	p2 := monitorCoresAssigned(ts2)
	defer p2.Done()
	// Run Hotel job
	go func() {
		runOps(ts2, ji, runHotel, rs2)
		done <- true
	}()
	// Wait for hotel jobs to set up.
	<-hotelJobs[0].ready
	// Run MR job
	go func() {
		runOps(ts1, mrapps, runMR, rs1)
		done <- true
	}()
	// Wait for MR jobs to set up.
	<-mrjobs[0].ready
	// Kick off MR jobs.
	mrjobs[0].ready <- true
	//	// Sleep for a bit
	//	time.Sleep(70 * time.Second)
	// Kick off hotel jobs
	hotelJobs[0].ready <- true
	// Wait for both jobs to finish.
	<-done
	<-done
	printResultSummary(rs1)
	hotelJobs[0].lg.Stats()
	ts1.Shutdown()
	ts2.Shutdown()
}

// XXX Old realm balance benchmark involving KV & MR.
// Start a realm with a long-running BE mr job. Then, start a realm with a kv
// job. In phases, ramp the kv job's CPU utilization up and down, and watch the
// realm-level software balance resource requests across realms.
func TestKVMRRRB(t *testing.T) {
	done := make(chan bool)
	// Find the total number of cores available for spinners across all machines.
	ts := test.MakeTstateAll(t)
	countNClusterCores(ts)
	// Structures for mr
	ts1 := test.MakeTstateRealm(t, ts.RealmId())
	rs1 := benchmarks.MakeResults(1, benchmarks.E2E)
	// Structure for kv
	ts2 := test.MakeTstateRealm(t, REALM2)
	rs2 := benchmarks.MakeResults(1, benchmarks.E2E)
	// Prep MR job
	mrjobs, mrapps := makeNMRJobs(ts1, 1, MR_APP)
	// Prep KV job
	nclerks := []int{N_CLERK}
	kvjobs, ji := makeNKVJobs(ts2, 1, N_KVD, 0, nclerks, nil, CLERK_DURATION, proc.Tcore(KVD_NCORE), proc.Tcore(CLERK_NCORE), KV_AUTO, REDIS_ADDR)
	p1 := monitorCoresAssigned(ts1)
	defer p1.Done()
	p2 := monitorCoresAssigned(ts2)
	defer p2.Done()
	// Run KV job
	go func() {
		runOps(ts2, ji, runKV, rs2)
		done <- true
	}()
	// Wait for KV jobs to set up.
	<-kvjobs[0].ready
	// Run MR job
	go func() {
		runOps(ts1, mrapps, runMR, rs1)
		done <- true
	}()
	// Wait for MR jobs to set up.
	<-mrjobs[0].ready
	// Kick off MR jobs.
	mrjobs[0].ready <- true
	// Sleep for a bit
	time.Sleep(70 * time.Second)
	// Kick off KV jobs
	kvjobs[0].ready <- true
	// Wait for both jobs to finish.
	<-done
	<-done
	printResultSummary(rs1)
	printResultSummary(rs2)
	ts1.Shutdown()
	ts2.Shutdown()
}

func testWww(ts *test.Tstate, sigmaos bool) {
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	if sigmaos {
		countNClusterCores(ts)
		maybePregrowRealm(ts)
	}
	db.DPrintf(db.ALWAYS, "Running with %d clients", N_CLNT)
	jobs, ji := makeWwwJobs(ts, sigmaos, 1, proc.Tcore(WWWD_NCORE), WWWD_REQ_TYPE, N_TRIALS, N_CLNT, N_CLNT_REQ, WWWD_REQ_DELAY)
	// XXX Clean this up/hide this somehow.
	go func() {
		for _, j := range jobs {
			// Wait until ready
			<-j.ready
			// Ack to allow the job to proceed.
			j.ready <- true
		}
	}()
	if sigmaos {
		p := monitorCoresAssigned(ts)
		defer p.Done()
	}
	runOps(ts, ji, runWww, rs)
	printResultSummary(rs)
	if sigmaos {
		ts.Shutdown()
	}
}

func TestWwwSigmaos(t *testing.T) {
	ts := test.MakeTstateAll(t)
	testWww(ts, true)
}

func TestWwwK8s(t *testing.T) {
	ts := test.MakeTstateAll(t)
	testWww(ts, false)
}

func testHotel(ts *test.Tstate, sigmaos bool, fn hotelFn) {
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	if sigmaos {
		countNClusterCores(ts)
		maybePregrowRealm(ts)
	}
	jobs, ji := makeHotelJobs(ts, sigmaos, proc.Tcore(HOTEL_NCORE), HOTEL_DUR, HOTEL_MAX_RPS, fn)
	// XXX Clean this up/hide this somehow.
	go func() {
		for _, j := range jobs {
			// Wait until ready
			<-j.ready
			// Ack to allow the job to proceed.
			j.ready <- true
		}
	}()
	if sigmaos {
		p := monitorCoresAssigned(ts)
		defer p.Done()
	}
	runOps(ts, ji, runHotel, rs)
	jobs[0].lg.Stats()
	//	printResultSummary(rs)
	if sigmaos {
		ts.Shutdown()
	}
}

func TestHotelSigmaosSearch(t *testing.T) {
	ts := test.MakeTstateAll(t)
	testHotel(ts, true, func(wc *hotel.WebClnt, r *rand.Rand) {
		hotel.RandSearchReq(wc, r)
	})
}

func TestHotelK8sSearch(t *testing.T) {
	ts := test.MakeTstateAll(t)
	testHotel(ts, false, func(wc *hotel.WebClnt, r *rand.Rand) {
		hotel.RandSearchReq(wc, r)
	})
}

func TestHotelSigmaosAll(t *testing.T) {
	ts := test.MakeTstateAll(t)
	testHotel(ts, true, func(wc *hotel.WebClnt, r *rand.Rand) {
		hotel.RunDSB(ts.T, 1, wc, r)
	})
}

func TestHotelK8sAll(t *testing.T) {
	ts := test.MakeTstateAll(t)
	testHotel(ts, false, func(wc *hotel.WebClnt, r *rand.Rand) {
		hotel.RunDSB(ts.T, 1, wc, r)
	})
}
