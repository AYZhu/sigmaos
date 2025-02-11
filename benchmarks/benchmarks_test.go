package benchmarks_test

import (
	"flag"
	"math/rand"
	"net/rpc"
	"path"
	"strconv"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"sigmaos/benchmarks"
	db "sigmaos/debug"
	"sigmaos/hotel"
	"sigmaos/linuxsched"
	"sigmaos/perf"
	"sigmaos/proc"
	"sigmaos/rpcbench"
	"sigmaos/rpcclnt"
	"sigmaos/scheddclnt"
	sp "sigmaos/sigmap"
	"sigmaos/test"
	"os/exec"
)

const (
	REALM_BASENAME sp.Trealm = "benchrealm"
	//	REALM1                   = "arielck" // NOTE: Set this as realm name to take cold-start into account.
	REALM1 = REALM_BASENAME + "1"
	REALM2 = REALM_BASENAME + "2"

	MR_K8S_INIT_PORT int = 32585

	HOSTTMP string = "/tmp/"
)

// Parameters
var N_TRIALS int
var N_THREADS int
var PREWARM_REALM bool
var MR_APP string
var KV_AUTO string
var N_KVD int
var N_CLERK int
var CLERK_DURATION string
var CLERK_MCPU int
var N_CLNT int
var N_CLNT_REQ int
var KVD_MCPU int
var WWWD_MCPU int
var WWWD_REQ_TYPE string
var WWWD_REQ_DELAY time.Duration
var HOTEL_NCACHE int
var HOTEL_CACHE_MCPU int
var N_HOTEL int
var HOTEL_IMG_SZ_MB int
var HOTEL_CACHE_AUTOSCALE bool
var CACHE_TYPE string
var CACHE_GC bool
var BLOCK_MEM string
var N_REALM int

// XXX Remove
var MEMCACHED_ADDRS string
var HTTP_URL string
var DURATION time.Duration
var MAX_RPS int
var HOTEL_DURS string
var HOTEL_MAX_RPS string
var SOCIAL_NETWORK_DURS string
var SOCIAL_NETWORK_MAX_RPS string
var SOCIAL_NETWORK_READ_ONLY bool
var RPCBENCH_MCPU int
var RPCBENCH_DURS string
var RPCBENCH_MAX_RPS string
var IMG_RESIZE_INPUT_PATH string
var N_IMG_RESIZE_JOBS int
var N_IMG_RESIZE_INPUTS_PER_JOB int
var IMG_RESIZE_MCPU int
var SLEEP time.Duration
var REDIS_ADDR string
var N_PROC int
var MCPU int
var MAT_SIZE int
var CONTENDERS_FRAC float64
var GO_MAX_PROCS int
var MAX_PARALLEL int
var K8S_ADDR string
var K8S_LEADER_NODE_IP string
var K8S_JOB_NAME string
var S3_RES_DIR string

// Read & set the proc version.
func init() {
	flag.IntVar(&N_REALM, "nrealm", 2, "Number of realms (relevant to BE balance benchmarks).")
	flag.IntVar(&N_TRIALS, "ntrials", 1, "Number of trials.")
	flag.IntVar(&N_THREADS, "nthreads", 1, "Number of threads.")
	flag.BoolVar(&PREWARM_REALM, "prewarm_realm", false, "Pre-warm realm, starting a BE and an LC uprocd on every machine in the cluster.")
	flag.StringVar(&MR_APP, "mrapp", "mr-wc-wiki1.8G.yml", "Name of mr yaml file.")
	flag.StringVar(&KV_AUTO, "kvauto", "manual", "KV auto-growing/shrinking.")
	flag.IntVar(&N_KVD, "nkvd", 1, "Number of kvds.")
	flag.IntVar(&N_CLERK, "nclerk", 1, "Number of clerks.")
	flag.IntVar(&N_CLNT, "nclnt", 1, "Number of clients.")
	flag.IntVar(&N_CLNT_REQ, "nclnt_req", 1, "Number of request each client makes.")
	flag.StringVar(&CLERK_DURATION, "clerk_dur", "90s", "Clerk duration.")
	flag.IntVar(&CLERK_MCPU, "clerk_mcpu", 1000, "Clerk mCPU")
	flag.IntVar(&KVD_MCPU, "kvd_mcpu", 2000, "KVD mCPU")
	flag.IntVar(&WWWD_MCPU, "wwwd_mcpu", 2000, "WWWD mCPU")
	flag.StringVar(&WWWD_REQ_TYPE, "wwwd_req_type", "compute", "WWWD request type [compute, dummy, io].")
	flag.DurationVar(&WWWD_REQ_DELAY, "wwwd_req_delay", 500*time.Millisecond, "Average request delay.")
	flag.DurationVar(&SLEEP, "sleep", 0*time.Millisecond, "Sleep length.")
	flag.IntVar(&HOTEL_NCACHE, "hotel_ncache", 1, "Hotel ncache")
	flag.IntVar(&HOTEL_CACHE_MCPU, "hotel_cache_mcpu", 2000, "Hotel cache mcpu")
	flag.IntVar(&HOTEL_IMG_SZ_MB, "hotel_img_sz_mb", 0, "Hotel image data size in megabytes.")
	flag.IntVar(&N_HOTEL, "nhotel", 80, "Number of hotels in the dataset.")
	flag.BoolVar(&HOTEL_CACHE_AUTOSCALE, "hotel_cache_autoscale", false, "Autoscale hotel cache")
	flag.StringVar(&CACHE_TYPE, "cache_type", "cached", "Hotel cache type (kvd or cached).")
	flag.BoolVar(&CACHE_GC, "cache_gc", false, "Turn hotel cache GC on (true) or off (false).")
	flag.StringVar(&BLOCK_MEM, "block_mem", "0MB", "Amount of physical memory to block on every machine.")
	flag.StringVar(&MEMCACHED_ADDRS, "memcached", "", "memcached server addresses (comma-separated).")
	flag.StringVar(&HTTP_URL, "http_url", "http://x.x.x.x", "HTTP url.")
	flag.DurationVar(&DURATION, "duration", 10*time.Second, "Duration.")
	flag.IntVar(&MAX_RPS, "max_rps", 1000, "Max requests per second.")
	flag.StringVar(&HOTEL_DURS, "hotel_dur", "10s", "Hotel benchmark load generation duration (comma-separated for multiple phases).")
	flag.StringVar(&HOTEL_MAX_RPS, "hotel_max_rps", "1000", "Max requests/second for hotel bench (comma-separated for multiple phases).")
	flag.StringVar(&SOCIAL_NETWORK_DURS, "sn_dur", "10s", "Social network benchmark load generation duration (comma-separated for multiple phases).")
	flag.StringVar(&SOCIAL_NETWORK_MAX_RPS, "sn_max_rps", "1000", "Max requests/second for social network bench (comma-separated for multiple phases).")
	flag.BoolVar(&SOCIAL_NETWORK_READ_ONLY, "sn_read_only", false, "send read only cases in social network bench")
	flag.StringVar(&RPCBENCH_DURS, "rpcbench_dur", "10s", "RPCBench benchmark load generation duration (comma-separated for multiple phases).")
	flag.StringVar(&RPCBENCH_MAX_RPS, "rpcbench_max_rps", "1000", "Max requests/second for rpc bench (comma-separated for multiple phases).")
	flag.IntVar(&RPCBENCH_MCPU, "rpcbench_mcpu", 3000, "RPCbench mCPU")
	flag.StringVar(&K8S_ADDR, "k8saddr", "", "Kubernetes frontend service address (only for hotel benchmarking for the time being).")
	flag.StringVar(&K8S_LEADER_NODE_IP, "k8sleaderip", "", "Kubernetes leader node ip.")
	flag.StringVar(&K8S_JOB_NAME, "k8sjobname", "thumbnail-benchrealm1", "Name of k8s job")
	flag.StringVar(&S3_RES_DIR, "s3resdir", "", "Results dir in s3.")
	flag.StringVar(&REDIS_ADDR, "redisaddr", "", "Redis server address")
	flag.IntVar(&N_PROC, "nproc", 1, "Number of procs per trial.")
	flag.IntVar(&MCPU, "mcpu", 1000, "Generic proc test MCPU")
	flag.IntVar(&MAT_SIZE, "matrixsize", 4000, "Size of matrix.")
	flag.Float64Var(&CONTENDERS_FRAC, "contenders", 4000, "Fraction of cores which should be taken up by contending procs.")
	flag.IntVar(&GO_MAX_PROCS, "gomaxprocs", int(linuxsched.NCores), "Go maxprocs setting for procs to be spawned.")
	flag.IntVar(&MAX_PARALLEL, "max_parallel", 1, "Max amount of parallelism.")
	flag.StringVar(&IMG_RESIZE_INPUT_PATH, "imgresize_path", "9ps3/img/6.jpg", "Path of img resize input file.")
	flag.IntVar(&N_IMG_RESIZE_JOBS, "n_imgresize", 10, "Number of img resize jobs.")
	flag.IntVar(&N_IMG_RESIZE_INPUTS_PER_JOB, "n_imgresize_per", 1, "Number of img resize inputs per job.")
	flag.IntVar(&IMG_RESIZE_MCPU, "imgresize_mcpu", 100, "MCPU for img resize worker.")
}

// ========== Common parameters ==========
const (
	OUT_DIR = "name/out_dir"
)

// Test how long it takes to init a semaphore.
func TestMicroInitSemaphore(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	rs := benchmarks.MakeResults(N_TRIALS, benchmarks.OPS)
	makeOutDir(ts1)
	_, is := makeNSemaphores(ts1, N_TRIALS)
	runOps(ts1, is, initSemaphore, rs)
	printResultSummary(rs)
	rmOutDir(ts1)
	rootts.Shutdown()
}

// Test how long it takes to up a semaphore.
func TestMicroUpSemaphore(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	rs := benchmarks.MakeResults(N_TRIALS, benchmarks.OPS)
	makeOutDir(ts1)
	_, is := makeNSemaphores(ts1, N_TRIALS)
	// Init semaphores first.
	for _, i := range is {
		initSemaphore(ts1, i)
	}
	runOps(ts1, is, upSemaphore, rs)
	printResultSummary(rs)
	rmOutDir(ts1)
	rootts.Shutdown()
}

// Test how long it takes to down a semaphore.
func TestMicroDownSemaphore(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	rs := benchmarks.MakeResults(N_TRIALS, benchmarks.OPS)
	makeOutDir(ts1)
	_, is := makeNSemaphores(ts1, N_TRIALS)
	// Init semaphores first.
	for _, i := range is {
		initSemaphore(ts1, i)
		upSemaphore(ts1, i)
	}
	runOps(ts1, is, downSemaphore, rs)
	printResultSummary(rs)
	rmOutDir(ts1)
	rootts.Shutdown()
}

// Test how long it takes to Spawn, run, and WaitExit a 5ms proc.
func TestMicroSpawnWaitStart(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	if PREWARM_REALM {
		warmupRealm(ts1)
	}
	rs := benchmarks.MakeResults(N_TRIALS, benchmarks.OPS)
	makeOutDir(ts1)
	ps, _ := makeNProcs(N_TRIALS, "sleeper", []string{"10000us", OUT_DIR}, nil, proc.Tmcpu(MCPU))
	runOps(ts1, []interface{}{ps}, spawnWaitStartProcs, rs)
	printResultSummary(rs)
	rmOutDir(ts1)
	rootts.Shutdown()
}

// Test how long it takes to Spawn, run, and WaitExit a 5ms proc.
func TestMicroSpawnWaitExit5msSleeper(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	if PREWARM_REALM {
		warmupRealm(ts1)
	}
	rs := benchmarks.MakeResults(N_TRIALS, benchmarks.OPS)
	makeOutDir(ts1)
	_, ps := makeNProcs(N_TRIALS, "sleeper", []string{"5000us", OUT_DIR}, nil, 1)
	runOps(ts1, ps, runProc, rs)
	printResultSummary(rs)
	rmOutDir(ts1)
	rootts.Shutdown()
}

// Test the throughput of spawning procs.
func TestMicroSpawnBurstTpt(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	rs := benchmarks.MakeResults(N_TRIALS, benchmarks.OPS)
	db.DPrintf(db.ALWAYS, "SpawnBursting %v procs (ncore=%v) with max parallelism %v", N_PROC, MCPU, MAX_PARALLEL)
	ps, _ := makeNProcs(N_PROC, "sleeper", []string{"0s", ""}, nil, proc.Tmcpu(MCPU))
	runOps(ts1, []interface{}{ps}, spawnBurstWaitStartProcs, rs)
	printResultSummary(rs)
	waitExitProcs(ts1, ps)
	rootts.Shutdown()
}

// Test the throughput of spawning procs.
func TestMicroHTTPLoadGen(t *testing.T) {
	RunHTTPLoadGen(HTTP_URL, DURATION, MAX_RPS)
}

func TestAppMR(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	if PREWARM_REALM {
		warmupRealm(ts1)
	}
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	p := makeRealmPerf(ts1)
	defer p.Done()
	jobs, apps := makeNMRJobs(ts1, p, 1, MR_APP)
	go func() {
		for _, j := range jobs {
			// Wait until ready
			<-j.ready
			// Ack to allow the job to proceed.
			j.ready <- true
		}
	}()
	monitorCPUUtil(ts1, p)
	runOps(ts1, apps, runMR, rs)
	printResultSummary(rs)
	rootts.Shutdown()
}

func runKVTest(t *testing.T, nReplicas int) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	p := makeRealmPerf(ts1)
	defer p.Done()
	nclerks := []int{N_CLERK}
	db.DPrintf(db.ALWAYS, "Running with %v clerks", N_CLERK)
	jobs, ji := makeNKVJobs(ts1, 1, N_KVD, nReplicas, nclerks, nil, CLERK_DURATION, proc.Tmcpu(KVD_MCPU), proc.Tmcpu(CLERK_MCPU), KV_AUTO, REDIS_ADDR)
	go func() {
		for _, j := range jobs {
			// Wait until ready
			<-j.ready
			// Ack to allow the job to proceed.
			j.ready <- true
		}
	}()
	monitorCPUUtil(ts1, p)
	db.DPrintf(db.TEST, "runOps")
	runOps(ts1, ji, runKV, rs)
	printResultSummary(rs)
	rootts.Shutdown()
}

func TestAppKVUnrepl(t *testing.T) {
	runKVTest(t, 0)
}

func TestAppKVRepl(t *testing.T) {
	runKVTest(t, 3)
}

func TestAppCached(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	p := makeRealmPerf(ts1)
	defer p.Done()
	const NKEYS = 100
	db.DPrintf(db.ALWAYS, "Running with %v clerks", N_CLERK)
	jobs, ji := makeNCachedJobs(ts1, 1, NKEYS, N_KVD, N_CLERK, CLERK_DURATION, proc.Tmcpu(CLERK_MCPU), proc.Tmcpu(KVD_MCPU))
	go func() {
		for _, j := range jobs {
			// Wait until ready
			<-j.ready
			// Ack to allow the job to proceed.
			j.ready <- true
		}
	}()
	monitorCPUUtil(ts1, p)
	runOps(ts1, ji, runCached, rs)
	printResultSummary(rs)
	rootts.Shutdown()
}

// Burst a bunch of spinning procs, and see how long it takes for all of them
// to start.
func TestRealmBurst(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	ncores := countClusterCores(rootts) - 1
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	makeOutDir(ts1)
	// Find the total number of cores available for spinners across all machines.
	// We need to get this in order to find out how many spinners to start.
	db.DPrintf(db.ALWAYS, "Bursting %v spinning procs", ncores)
	ps, _ := makeNProcs(int(ncores), "spinner", []string{OUT_DIR}, nil, 1)
	p := makeRealmPerf(ts1)
	defer p.Done()
	monitorCPUUtil(ts1, p)
	runOps(ts1, []interface{}{ps}, spawnBurstWaitStartProcs, rs)
	printResultSummary(rs)
	evictProcs(ts1, ps)
	rmOutDir(ts1)
	rootts.Shutdown()
}

func TestLambdaBurst(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	makeOutDir(ts1)
	// Find the total number of cores available for spinners across all machines.
	// We need to get this in order to find out how many spinners to start.
	N_LAMBDAS := 720
	db.DPrintf(db.ALWAYS, "Invoking %v lambdas", N_LAMBDAS)
	ss, is := makeNSemaphores(ts1, N_LAMBDAS)
	// Init semaphores first.
	for _, i := range is {
		initSemaphore(ts1, i)
	}
	runOps(ts1, []interface{}{ss}, invokeWaitStartLambdas, rs)
	printResultSummary(rs)
	rmOutDir(ts1)
	rootts.Shutdown()
}

func TestLambdaInvokeWaitStart(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	rs := benchmarks.MakeResults(720, benchmarks.E2E)
	makeOutDir(ts1)
	N_LAMBDAS := 640
	db.DPrintf(db.ALWAYS, "Invoking %v lambdas", N_LAMBDAS)
	_, is := makeNSemaphores(ts1, N_LAMBDAS)
	// Init semaphores first.
	for _, i := range is {
		initSemaphore(ts1, i)
	}
	runOps(ts1, is, invokeWaitStartOneLambda, rs)
	printResultSummary(rs)
	rmOutDir(ts1)
	rootts.Shutdown()
}

// Start a realm with a long-running BE mr job. Then, start a realm with an LC
// hotel job. In phases, ramp the hotel job's CPU utilization up and down, and
// watch the realm-level software balance resource requests across realms.
func TestRealmBalanceMRHotel(t *testing.T) {
	done := make(chan bool)
	rootts := test.MakeTstateWithRealms(t)
	blockers := blockMem(rootts, BLOCK_MEM)
	// Structures for mr
	ts1 := test.MakeRealmTstate(rootts, REALM2)
	rs1 := benchmarks.MakeResults(1, benchmarks.E2E)
	p1 := makeRealmPerf(ts1)
	defer p1.Done()
	// Structure for hotel
	ts2 := test.MakeRealmTstate(rootts, REALM1)
	rs2 := benchmarks.MakeResults(1, benchmarks.E2E)
	p2 := makeRealmPerf(ts2)
	defer p2.Done()
	// Prep MR job
	mrjobs, mrapps := makeNMRJobs(ts1, p1, 1, MR_APP)
	// Prep Hotel job
	hotelJobs, ji := makeHotelJobs(ts2, p2, true, HOTEL_DURS, HOTEL_MAX_RPS, HOTEL_NCACHE, CACHE_TYPE, proc.Tmcpu(HOTEL_CACHE_MCPU), func(wc *hotel.WebClnt, r *rand.Rand) {
		//		hotel.RunDSB(ts2.T, 1, wc, r)
		err := hotel.RandSearchReq(wc, r)
		assert.Nil(t, err, "SearchReq %v", err)
	})
	// Monitor cores assigned to MR.
	monitorCPUUtil(ts1, p1)
	// Monitor cores assigned to Hotel.
	monitorCPUUtil(ts2, p2)
	// Run Hotel job
	go func() {
		runOps(ts2, ji, runHotel, rs2)
		done <- true
	}()
	// Wait for hotel jobs to set up.
	<-hotelJobs[0].ready
	db.DPrintf(db.TEST, "Hotel setup done.")
	// Run MR job
	go func() {
		runOps(ts1, mrapps, runMR, rs1)
		done <- true
	}()
	// Wait for MR jobs to set up.
	<-mrjobs[0].ready
	db.DPrintf(db.TEST, "MR setup done.")
	db.DPrintf(db.TEST, "Setup phase done.")
	if N_CLNT > 1 {
		// Wait for hotel clients to start up on other machines.
		db.DPrintf(db.ALWAYS, "Leader waiting for clnts")
		waitForClnts(rootts, N_CLNT)
		db.DPrintf(db.ALWAYS, "Leader done waiting for clnts")
	}
	db.DPrintf(db.TEST, "Done waiting for hotel clnts.")
	// Kick off MR jobs.
	mrjobs[0].ready <- true
	// Sleep for a bit
	time.Sleep(SLEEP)
	// Kick off hotel jobs
	hotelJobs[0].ready <- true
	// Wait for both jobs to finish.
	<-done
	<-done
	db.DPrintf(db.TEST, "MR and Hotel done.")
	printResultSummary(rs1)
	time.Sleep(20 * time.Second)
	evictMemBlockers(rootts, blockers)
	rootts.Shutdown()
}

// Start a realm with a long-running BE mr job. Then, start a realm with an LC
// hotel job. In phases, ramp the hotel job's CPU utilization up and down, and
// watch the realm-level software balance resource requests across realms.
func TestRealmBalanceMRMR(t *testing.T) {
	done := make(chan bool)
	rootts := test.MakeTstateWithRealms(t)
	tses := make([]*test.RealmTstate, N_REALM)
	rses := make([]*benchmarks.Results, N_REALM)
	ps := make([]*perf.Perf, N_REALM)
	mrjobs := make([][]*MRJobInstance, N_REALM)
	mrapps := make([][]interface{}, N_REALM)
	// Create structures for MR jobs.
	for i := range tses {
		tses[i] = test.MakeRealmTstate(rootts, sp.Trealm(REALM_BASENAME.String()+strconv.Itoa(i+1)))
		rses[i] = benchmarks.MakeResults(1, benchmarks.E2E)
		ps[i] = makeRealmPerf(tses[i])
		defer ps[i].Done()
		mrjob, mrapp := makeNMRJobs(tses[i], ps[i], 1, MR_APP)
		mrjobs[i] = mrjob
		mrapps[i] = mrapp
	}
	// Start CPU utilization monitoring.
	for i := range tses {
		monitorCPUUtil(tses[i], ps[i])
	}
	// Initialize MR jobs.
	for i := range tses {
		// Start MR job initialization.
		go func(ts *test.RealmTstate, mrapp []interface{}, rs *benchmarks.Results) {
			runOps(ts, mrapp, runMR, rs)
			done <- true
		}(tses[i], mrapps[i], rses[i])
		// Wait for MR job to set up.
		<-mrjobs[i][0].ready
	}
	// Start jobs running, with a small delay between each job start.
	for i := range tses {
		// Kick off MR jobs.
		mrjobs[i][0].ready <- true
		db.DPrintf(db.TEST, "Start MR job %v", i+1)
		// Sleep for a bit before starting the next job
		time.Sleep(SLEEP)
	}
	// Wait for both jobs to finish.
	for i := range tses {
		<-done
		db.DPrintf(db.TEST, "Done MR job %v", i+1)
	}
	printResultSummary(rses[0])
	rootts.Shutdown()
}

// Old realm balance benchmark involving KV & MR.
// Start a realm with a long-running BE mr job. Then, start a realm with a kv
// job. In phases, ramp the kv job's CPU utilization up and down, and watch the
// realm-level software balance resource requests across realms.
func TestKVMRRRB(t *testing.T) {
	done := make(chan bool)
	rootts := test.MakeTstateWithRealms(t)
	// Structures for mr
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	rs1 := benchmarks.MakeResults(1, benchmarks.E2E)
	p1 := makeRealmPerf(ts1)
	defer p1.Done()
	// Structure for kv
	ts2 := test.MakeRealmTstate(rootts, REALM2)
	rs2 := benchmarks.MakeResults(1, benchmarks.E2E)
	p2 := makeRealmPerf(ts2)
	defer p2.Done()
	// Prep MR job
	mrjobs, mrapps := makeNMRJobs(ts1, p1, 1, MR_APP)
	// Prep KV job
	nclerks := []int{N_CLERK}
	kvjobs, ji := makeNKVJobs(ts2, 1, N_KVD, 0, nclerks, nil, CLERK_DURATION, proc.Tmcpu(KVD_MCPU), proc.Tmcpu(CLERK_MCPU), KV_AUTO, REDIS_ADDR)
	monitorCPUUtil(ts1, p1)
	monitorCPUUtil(ts2, p2)
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
	rootts.Shutdown()
}

func testWww(t *testing.T, sigmaos bool) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	db.DPrintf(db.ALWAYS, "Running with %d clients", N_CLNT)
	jobs, ji := makeWwwJobs(ts1, sigmaos, 1, proc.Tmcpu(WWWD_MCPU), WWWD_REQ_TYPE, N_TRIALS, N_CLNT, N_CLNT_REQ, WWWD_REQ_DELAY)
	go func() {
		for _, j := range jobs {
			// Wait until ready
			<-j.ready
			// Ack to allow the job to proceed.
			j.ready <- true
		}
	}()
	if sigmaos {
		p := makeRealmPerf(ts1)
		defer p.Done()
		monitorCPUUtil(ts1, p)
	}
	runOps(ts1, ji, runWww, rs)
	printResultSummary(rs)
	if sigmaos {
		rootts.Shutdown()
	}
}

func TestWwwSigmaos(t *testing.T) {
	testWww(t, true)
}

func TestWwwK8s(t *testing.T) {
	testWww(t, false)
}

func testRPCBench(rootts *test.Tstate, ts1 *test.RealmTstate, p *perf.Perf, fn rpcbenchFn) {
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	jobs, ji := makeRPCBenchJobs(ts1, p, proc.Tmcpu(RPCBENCH_MCPU), RPCBENCH_DURS, RPCBENCH_MAX_RPS, fn)
	go func() {
		for _, j := range jobs {
			// Wait until ready
			<-j.ready
			if N_CLNT > 1 {
				// Wait for clients to start up on other machines.
				waitForClnts(rootts, N_CLNT)
			}
			// Ack to allow the job to proceed.
			j.ready <- true
		}
	}()
	p2 := makeRealmPerf(ts1)
	defer p2.Done()
	monitorCPUUtil(ts1, p2)
	runOps(ts1, ji, runRPCBench, rs)
	//	printResultSummary(rs)
	rootts.Shutdown()
}

func TestRPCBenchSigmaosSleep(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	testRPCBench(rootts, ts1, nil, func(c *rpcbench.Clnt) {
		err := c.Sleep(int64(SLEEP / time.Millisecond))
		assert.Nil(t, err, "Error sleep req: %v", err)
	})
}

func TestRPCBenchSigmaosJustCliSleep(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstateClnt(rootts, REALM1)
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	clientReady(rootts)
	jobs, ji := makeRPCBenchJobsCli(ts1, nil, proc.Tmcpu(RPCBENCH_MCPU), RPCBENCH_DURS, RPCBENCH_MAX_RPS, func(c *rpcbench.Clnt) {
		err := c.Sleep(int64(SLEEP / time.Millisecond))
		assert.Nil(t, err, "Error sleep req: %v", err)
	})
	go func() {
		for _, j := range jobs {
			// Wait until ready
			<-j.ready
			// Ack to allow the job to proceed.
			j.ready <- true
		}
	}()
	runOps(ts1, ji, runHotel, rs)
	//	printResultSummary(rs)
	//	jobs[0].requestK8sStats()
}

func testHotel(rootts *test.Tstate, ts1 *test.RealmTstate, p *perf.Perf, sigmaos bool, fn hotelFn) {
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	go func() {
		time.Sleep(20 * time.Second)
		if sts, err := rootts.GetDir(sp.WS_RUNQ_LC); err != nil || len(sts) > 0 {
			rootts.Shutdown()
			db.DFatalf("Error getdir ws err %v ws %v", err, sp.Names(sts))
		} else {
			db.DPrintf(db.ALWAYS, "Getdir contents %v : %v", sp.WS_RUNQ_LC, sp.Names(sts))
		}
	}()
	jobs, ji := makeHotelJobs(ts1, p, sigmaos, HOTEL_DURS, HOTEL_MAX_RPS, HOTEL_NCACHE, CACHE_TYPE, proc.Tmcpu(HOTEL_CACHE_MCPU), fn)
	go func() {
		for _, j := range jobs {
			// Wait until ready
			<-j.ready
			if N_CLNT > 1 {
				// Wait for clients to start up on other machines.
				db.DPrintf(db.ALWAYS, "Leader waiting for clnts")
				waitForClnts(rootts, N_CLNT)
				db.DPrintf(db.ALWAYS, "Leader done waiting for clnts")
			}
			// Ack to allow the job to proceed.
			j.ready <- true
		}
	}()
	if sigmaos {
		p := makeRealmPerf(ts1)
		defer p.Done()
		monitorCPUUtil(ts1, p)
	} else {
		p := makeRealmPerf(ts1)
		defer p.Done()
		monitorK8sCPUUtil(ts1, p, "hotel", "")
	}
	runOps(ts1, ji, runHotel, rs)
	//	printResultSummary(rs)
	if sigmaos {
		//		rootts.Shutdown()
	} else {
		jobs[0].requestK8sStats()
	}
}

// XXX Messy, get rid of this.
var reservec *rpcclnt.RPCClnt

func TestHotelSigmaosReserve(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	testHotel(rootts, ts1, nil, true, func(wc *hotel.WebClnt, r *rand.Rand) {
		err := hotel.RandCheckAvailabilityReq(reservec, r)
		assert.Nil(t, err, "Error reserve req: %v", err)
	})
}

func TestHotelSigmaosSearch(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	testHotel(rootts, ts1, nil, true, func(wc *hotel.WebClnt, r *rand.Rand) {
		err := hotel.RandSearchReq(wc, r)
		assert.Nil(t, err, "Error search req: %v", err)
	})
}

func TestHotelSigmaosJustCliSearch(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstateClnt(rootts, REALM1)
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	clientReady(rootts)
	// Sleep for a bit
	time.Sleep(SLEEP)
	if sts, err := rootts.GetDir(sp.WS_RUNQ_LC); err != nil || len(sts) > 0 {
		rootts.Shutdown()
		db.DFatalf("Error getdir ws err %v ws %v", err, sp.Names(sts))
	} else {
		db.DPrintf(db.ALWAYS, "Getdir contents %v : %v", sp.WS_RUNQ_LC, sp.Names(sts))
	}
	jobs, ji := makeHotelJobsCli(ts1, true, HOTEL_DURS, HOTEL_MAX_RPS, HOTEL_NCACHE, CACHE_TYPE, proc.Tmcpu(HOTEL_CACHE_MCPU), func(wc *hotel.WebClnt, r *rand.Rand) {
		err := hotel.RandSearchReq(wc, r)
		assert.Nil(t, err, "Error search req: %v", err)
	})
	go func() {
		for _, j := range jobs {
			// Wait until ready
			<-j.ready
			// Ack to allow the job to proceed.
			j.ready <- true
		}
	}()
	runOps(ts1, ji, runHotel, rs)
	//	printResultSummary(rs)
	//	jobs[0].requestK8sStats()
}

func TestHotelK8sJustCliSearch(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstateClnt(rootts, REALM1)
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	db.DPrintf(db.ALWAYS, "Clnt ready")
	clientReady(rootts)
	db.DPrintf(db.ALWAYS, "Clnt done waiting")
	jobs, ji := makeHotelJobsCli(ts1, false, HOTEL_DURS, HOTEL_MAX_RPS, HOTEL_NCACHE, CACHE_TYPE, proc.Tmcpu(HOTEL_CACHE_MCPU), func(wc *hotel.WebClnt, r *rand.Rand) {
		err := hotel.RandSearchReq(wc, r)
		assert.Nil(t, err, "Error search req: %v", err)
	})
	go func() {
		for _, j := range jobs {
			// Wait until ready
			<-j.ready
			// Ack to allow the job to proceed.
			j.ready <- true
		}
	}()
	runOps(ts1, ji, runHotel, rs)
	//	printResultSummary(rs)
	//	jobs[0].requestK8sStats()
}

func TestHotelK8sSearch(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	testHotel(rootts, ts1, nil, false, func(wc *hotel.WebClnt, r *rand.Rand) {
		err := hotel.RandSearchReq(wc, r)
		assert.Nil(t, err, "Error search req: %v", err)
	})
	downloadS3Results(rootts, path.Join("name/s3/~any/9ps3/", "hotelperf/k8s"), HOSTTMP+"sigmaos-perf")
}

func TestHotelK8sSearchCli(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	testHotel(rootts, ts1, nil, false, func(wc *hotel.WebClnt, r *rand.Rand) {
		hotel.RandSearchReq(wc, r)
	})
}

func TestHotelSigmaosAll(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	testHotel(rootts, ts1, nil, true, func(wc *hotel.WebClnt, r *rand.Rand) {
		hotel.RunDSB(rootts.T, 1, wc, r)
	})
}

func TestHotelK8sAll(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	testHotel(rootts, ts1, nil, false, func(wc *hotel.WebClnt, r *rand.Rand) {
		hotel.RunDSB(rootts.T, 1, wc, r)
	})
}

func TestMRK8s(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	assert.NotEqual(rootts.T, K8S_LEADER_NODE_IP, "", "Must pass k8s leader node ip")
	assert.NotEqual(rootts.T, S3_RES_DIR, "", "Must pass s3 reulst dir")
	if K8S_LEADER_NODE_IP == "" || S3_RES_DIR == "" {
		db.DPrintf(db.ALWAYS, "Skipping mr k8s")
		return
	}
	c := startK8sMR(rootts, k8sMRAddr(K8S_LEADER_NODE_IP, MR_K8S_INIT_PORT))
	waitK8sMR(rootts, c)
	downloadS3Results(rootts, path.Join("name/s3/~any/9ps3/", S3_RES_DIR), HOSTTMP+"sigmaos-perf")
}

func TestK8sMRMulti(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	assert.NotEqual(rootts.T, K8S_LEADER_NODE_IP, "", "Must pass k8s leader node ip")
	assert.NotEqual(rootts.T, S3_RES_DIR, "", "Must pass s3 result dir")
	if K8S_LEADER_NODE_IP == "" || S3_RES_DIR == "" {
		db.DPrintf(db.ALWAYS, "Skipping mr k8s")
		return
	}
	// Create realm structures.
	ts := make([]*test.RealmTstate, 0, N_REALM)
	ps := make([]*perf.Perf, 0, N_REALM)
	for i := 0; i < N_REALM; i++ {
		rName := sp.Trealm(REALM_BASENAME.String() + strconv.Itoa(i+1))
		db.DPrintf(db.TEST, "Create realm srtructs for %v", rName)
		ts = append(ts, test.MakeRealmTstate(rootts, rName))
		ps = append(ps, makeRealmPerf(ts[i]))
		defer ps[i].Done()
	}
	db.DPrintf(db.TEST, "Done creating realm srtructs")
	err := ts[0].MkDir(sp.K8S_SCRAPER, 0777)
	assert.Nil(rootts.T, err, "Error mkdir %v", err)
	// Start up the stat scraper procs.
	sdc := scheddclnt.MakeScheddClnt(ts[0].SigmaClnt.FsLib)
	nSchedd, err := sdc.Nschedd()
	ps2, _ := makeNProcs(nSchedd, "k8s-stat-scraper", []string{}, nil, proc.Tmcpu(1000*(linuxsched.NCores-1)))
	spawnBurstProcs(ts[0], ps2)
	waitStartProcs(ts[0], ps2)

	cs := make([]*rpc.Client, 0, N_REALM)
	for i := 0; i < N_REALM; i++ {
		rName := sp.Trealm(REALM_BASENAME.String() + strconv.Itoa(i+1))
		db.DPrintf(db.TEST, "Starting MR job for realm %v", rName)
		// Start the next k8s job.
		cs = append(cs, startK8sMR(rootts, k8sMRAddr(K8S_LEADER_NODE_IP, MR_K8S_INIT_PORT+i+1)))
		// Monitor cores assigned to this realm.
		//		monitorK8sCPUUtil(ts[i], ps[i], "mr", rName)
		monitorK8sCPUUtilScraperTS(ts[0], ps[i], "Guaranteed")
		// Sleep for a bit before starting the next job
		time.Sleep(SLEEP)
	}
	db.DPrintf(db.TEST, "Done starting MR jobs")
	for i, c := range cs {
		waitK8sMR(rootts, c)
		db.DPrintf(db.TEST, "MR job %v finished", i)
	}
	db.DPrintf(db.TEST, "Done waiting for MR jobs.")
	for i := 0; i < N_REALM; i++ {
		downloadS3ResultsRealm(
			rootts,
			path.Join("name/s3/~any/9ps3/", S3_RES_DIR+"-"+strconv.Itoa(i+1)),
			HOSTTMP+"sigmaos-perf",
			sp.Trealm(REALM_BASENAME.String()+strconv.Itoa(i+1)),
		)
	}
	db.DPrintf(db.TEST, "Done downloading results.")
}

func TestK8sBalanceHotelMR(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	// Structures for mr
	ts1 := test.MakeRealmTstate(rootts, REALM2)
	p1 := makeRealmPerf(ts1)
	defer p1.Done()
	// Structure for hotel
	ts2 := test.MakeRealmTstate(rootts, REALM1)
	p2 := makeRealmPerf(ts2)
	defer p2.Done()
	// Monitor cores assigned to MR.
	monitorK8sCPUUtil(ts1, p1, "mr", "")
	// Monitor cores assigned to Hotel.
	monitorK8sCPUUtil(ts2, p2, "hotel", "")
	assert.NotEqual(rootts.T, K8S_LEADER_NODE_IP, "", "Must pass k8s leader node ip")
	assert.NotEqual(rootts.T, S3_RES_DIR, "", "Must pass k8s leader node ip")
	db.DPrintf(db.TEST, "Starting hotel")
	done := make(chan bool)
	go func() {
		testHotel(rootts, ts2, nil, false, func(wc *hotel.WebClnt, r *rand.Rand) {
			hotel.RandSearchReq(wc, r)
		})
		done <- true
	}()
	db.DPrintf(db.TEST, "Starting mr")
	if K8S_LEADER_NODE_IP == "" || S3_RES_DIR == "" {
		db.DPrintf(db.ALWAYS, "Skipping mr k8s")
		return
	}
	c := startK8sMR(rootts, k8sMRAddr(K8S_LEADER_NODE_IP, MR_K8S_INIT_PORT))
	waitK8sMR(rootts, c)
	<-done
	db.DPrintf(db.TEST, "Downloading results")
	downloadS3Results(rootts, path.Join("name/s3/~any/9ps3/", S3_RES_DIR), HOSTTMP+"sigmaos-perf")
	downloadS3Results(rootts, path.Join("name/s3/~any/9ps3/", "hotelperf/k8s"), HOSTTMP+"sigmaos-perf")
}

func TestImgResize(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	if PREWARM_REALM {
		warmupRealm(ts1)
	}
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	p := makeRealmPerf(ts1)
	defer p.Done()
	jobs, apps := makeImgResizeJob(ts1, p, true, IMG_RESIZE_INPUT_PATH, N_IMG_RESIZE_JOBS, N_IMG_RESIZE_INPUTS_PER_JOB, proc.Tmcpu(IMG_RESIZE_MCPU))
	go func() {
		for _, j := range jobs {
			// Wait until ready
			<-j.ready
			// Ack to allow the job to proceed.
			j.ready <- true
		}
	}()
	monitorCPUUtil(ts1, p)
	runOps(ts1, apps, runImgResize, rs)
	printResultSummary(rs)
	rootts.Shutdown()
}

func TestK8sImgResize(t *testing.T) {
	rootts := test.MakeTstateWithRealms(t)
	ts1 := test.MakeRealmTstateClnt(rootts, sp.ROOTREALM)
	if PREWARM_REALM {
		warmupRealm(ts1)
	}
	sdc := scheddclnt.MakeScheddClnt(ts1.FsLib)
	nSchedd, err := sdc.Nschedd()
	assert.Nil(ts1.Ts.T, err, "Error nschedd %v", err)
	rs := benchmarks.MakeResults(1, benchmarks.E2E)
	p := makeRealmPerf(ts1)
	defer p.Done()
	err = ts1.MkDir(sp.K8S_SCRAPER, 0777)
	assert.Nil(ts1.Ts.T, err, "Error mkdir %v", err)
	// Start up the stat scraper procs.
	ps, _ := makeNProcs(nSchedd, "k8s-stat-scraper", []string{}, nil, proc.Tmcpu(1000*(linuxsched.NCores-1)))
	spawnBurstProcs(ts1, ps)
	waitStartProcs(ts1, ps)
	// NOte start time
	start := time.Now()
	// Monitor CPU utilization via the stat scraper procs.
	monitorK8sCPUUtilScraper(rootts, p, "BestEffort")
	exec.Command("kubectl", "apply", "-Rf", "/tmp/thumbnail.yaml").Start()
	for !k8sJobHasCompleted(K8S_JOB_NAME) {
		time.Sleep(500 * time.Millisecond)
	}
	rs.Append(time.Since(start), 1)
	printResultSummary(rs)
	evictProcs(ts1, ps)
	rootts.Shutdown()
}

func TestRealmBalanceSimpleImgResize(t *testing.T) {
	done := make(chan bool)
	rootts := test.MakeTstateWithRealms(t)
	blockers := blockMem(rootts, BLOCK_MEM)
	// Structures for BE image resize
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	rs1 := benchmarks.MakeResults(1, benchmarks.E2E)
	p1 := makeRealmPerf(ts1)
	defer p1.Done()
	// Structure for LC image resize
	ts2 := test.MakeRealmTstate(rootts, REALM2)
	rs2 := benchmarks.MakeResults(1, benchmarks.E2E)
	p2 := makeRealmPerf(ts2)
	defer p2.Done()
	// Prep resize jobs
	imgJobsBE, imgAppsBE := makeImgResizeJob(
		ts1, p1, true, IMG_RESIZE_INPUT_PATH, N_IMG_RESIZE_JOBS, N_IMG_RESIZE_INPUTS_PER_JOB, 0)
	imgJobsLC, imgAppsLC := makeImgResizeJob(
		ts2, p2, true, IMG_RESIZE_INPUT_PATH, N_IMG_RESIZE_JOBS, N_IMG_RESIZE_INPUTS_PER_JOB, proc.Tmcpu(IMG_RESIZE_MCPU))

	// Run image resize jobs
	go func() {
		runOps(ts1, imgAppsBE, runImgResize, rs1)
		done <- true
	}()
	go func() {
		runOps(ts2, imgAppsLC, runImgResize, rs2)
		done <- true
	}()
	// Wait for image resize jobs to set up.
	<-imgJobsBE[0].ready
	<-imgJobsLC[0].ready

	// Monitor cores for kernel procs
	monitorCPUUtil(ts1, p1)
	monitorCPUUtil(ts2, p2)
	db.DPrintf(db.TEST, "Image Resize setup done.")
	// Kick off image resize jobs.
	imgJobsBE[0].ready <- true
	// Sleep for a bit
	time.Sleep(5 * time.Second)
	// Kick off social network jobs
	imgJobsLC[0].ready <- true
	// Wait for both jobs to finish.
	<-done
	<-done
	db.DPrintf(db.TEST, "Image Resize Done.")
	printResultSummary(rs1)
	evictMemBlockers(rootts, blockers)
	rootts.Shutdown()
}

func TestRealmBalanceSocialNetworkImgResize(t *testing.T) {
	done := make(chan bool)
	rootts := test.MakeTstateWithRealms(t)
	blockers := blockMem(rootts, BLOCK_MEM)
	ts0 := test.MakeRealmTstateClnt(rootts, "rootrealm")
	p0 := makeRealmPerf(ts0)
	defer p0.Done()
	// Structures for image resize
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	rs1 := benchmarks.MakeResults(1, benchmarks.E2E)
	p1 := makeRealmPerf(ts1)
	defer p1.Done()
	// Structure for social network
	ts2 := test.MakeRealmTstate(rootts, REALM2)
	rs2 := benchmarks.MakeResults(1, benchmarks.E2E)
	p2 := makeRealmPerf(ts2)
	defer p2.Done()
	// Prep image resize job
	imgJobs, imgApps := makeImgResizeJob(
		ts1, p1, true, IMG_RESIZE_INPUT_PATH, N_IMG_RESIZE_JOBS, N_IMG_RESIZE_INPUTS_PER_JOB, 0)
	// Prep social network job
	snJobs, snApps := makeSocialNetworkJobs(ts2, p2, true, SOCIAL_NETWORK_READ_ONLY, SOCIAL_NETWORK_DURS, SOCIAL_NETWORK_MAX_RPS, 3)
	// Run social network job
	go func() {
		runOps(ts2, snApps, runSocialNetwork, rs2)
		done <- true
	}()
	// Wait for social network jobs to set up.
	<-snJobs[0].ready
	db.DPrintf(db.TEST, "Social Network setup done.")
	// Run image resize job
	go func() {
		runOps(ts1, imgApps, runImgResize, rs1)
		done <- true
	}()
	// Wait for image resize jobs to set up.
	<-imgJobs[0].ready
	// Monitor cores for kernel procs
	monitorCPUUtil(ts0, p0)
	monitorCPUUtil(ts1, p1)
	monitorCPUUtil(ts2, p2)
	db.DPrintf(db.TEST, "Image Resize setup done.")
	db.DPrintf(db.TEST, "Setup phase done.")
	// Kick off image resize jobs.
	imgJobs[0].ready <- true
	// Sleep for a bit
	time.Sleep(10 * time.Second)
	// Kick off social network jobs
	snJobs[0].ready <- true
	// Wait for both jobs to finish.
	<-done
	<-done
	db.DPrintf(db.TEST, "Image Resize and Social Network Done.")
	printResultSummary(rs1)
	time.Sleep(5 * time.Second)
	evictMemBlockers(rootts, blockers)
	rootts.Shutdown()
}

func TestK8sSocialNetworkImgResize(t *testing.T) {
	done := make(chan bool)
	rootts := test.MakeTstateWithRealms(t)
	blockers := blockMem(rootts, BLOCK_MEM)
	// make realm to run k8s scrapper
	ts0:= test.MakeRealmTstateClnt(rootts, sp.ROOTREALM)
	p0 := makeRealmPerf(ts0)
	defer p0.Done()
	if PREWARM_REALM {
		warmupRealm(ts0)
	}
	sdc := scheddclnt.MakeScheddClnt(ts0.SigmaClnt.FsLib)
	nSchedd, err := sdc.Nschedd()
	assert.Nil(ts0.Ts.T, err, "Error nschedd %v", err)
	rs0 := benchmarks.MakeResults(1, benchmarks.E2E)
	err = ts0.MkDir(sp.K8S_SCRAPER, 0777)
	assert.Nil(ts0.Ts.T, err, "Error mkdir %v", err)
	// Start up the stat scraper procs.
	//ps, _ := makeNProcs(nSchedd, "k8s-stat-scraper", []string{}, nil, proc.Tmcpu(1000*(linuxsched.NCores-1)))
	ps, _ := makeNProcs(nSchedd, "k8s-stat-scraper", []string{}, nil, 0)
	spawnBurstProcs(ts0, ps)
	waitStartProcs(ts0, ps)
	// Structures for image resize
	ts1 := test.MakeRealmTstate(rootts, REALM1)
	//rs1 := benchmarks.MakeResults(1, benchmarks.E2E)
	p1 := makeRealmPerf(ts1)
	defer p1.Done()
	// Structure for social network
	ts2 := test.MakeRealmTstate(rootts, REALM2)
	rs2 := benchmarks.MakeResults(1, benchmarks.E2E)
	p2 := makeRealmPerf(ts2)
	defer p2.Done()
	// Prep image resize job

	// Prep social network job
	snJobs, snApps := makeSocialNetworkJobs(ts2, p2, false, SOCIAL_NETWORK_READ_ONLY, SOCIAL_NETWORK_DURS, SOCIAL_NETWORK_MAX_RPS, 3)
	// Monitor cores assigned to image resize.
	// NOte start time
	start := time.Now()
	// Run social network job
	go func() {
		runOps(ts2, snApps, runSocialNetwork, rs2)
		snJobs[0].requestK8sStats()
		done <- true
	}()
	// Wait for social network jobs to set up.
	<-snJobs[0].ready
	db.DPrintf(db.TEST, "Social Network setup done.")
	// Monitor CPU utilization via the stat scraper procs.
	monitorK8sCPUUtilScraper(rootts, p2, "Burstable")
	monitorK8sCPUUtilScraper(rootts, p1, "BestEffort")
	// Run image resize job
	exec.Command("kubectl", "apply", "-Rf", "/tmp/thumbnail-heavy/").Start()
	// Wait for image resize jobs to set up.
	db.DPrintf(db.TEST, "Setup phase done.")
	// Kick off image resize jobs.
	// Sleep for a bit
	time.Sleep(5 * time.Second)
	// Kick off social network jobs
	snJobs[0].ready <- true
	// Wait for both jobs to finish.
	<-done
	db.DPrintf(db.TEST, "Downloading results")
	downloadS3Results(rootts, path.Join("name/s3/~any/9ps3/", "social-network-perf/k8s"), HOSTTMP+"sigmaos-perf")
	for !(k8sJobHasCompleted("thumbnail1-benchrealm1") && k8sJobHasCompleted("thumbnail2-benchrealm1") && 
			k8sJobHasCompleted("thumbnail3-benchrealm1")&&k8sJobHasCompleted("thumbnail4-benchrealm1")) {
		time.Sleep(500 * time.Millisecond)
	}
	rs0.Append(time.Since(start), 1)
	printResultSummary(rs0)
	evictProcs(ts0, ps)
	time.Sleep(10 * time.Second)
	evictMemBlockers(rootts, blockers)
	rootts.Shutdown()
}
