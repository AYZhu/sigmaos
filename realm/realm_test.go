package realm_test

import (
	"flag"
	"runtime/debug"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"sigmaos/config"
	db "sigmaos/debug"
	"sigmaos/fslib"
	"sigmaos/linuxsched"
	"sigmaos/proc"
	"sigmaos/procclnt"
	"sigmaos/realm"
	sp "sigmaos/sigmap"
)

const (
	SLEEP_TIME_MS = 3000
)

var version string

// Read & set the proc version.
func init() {
	flag.StringVar(&version, "version", "none", "version")
}

type Tstate struct {
	t        *testing.T
	e        *realm.TestEnv
	cfg      *realm.RealmConfig
	realmFsl *fslib.FsLib
	*config.ConfigClnt
	coreGroupsPerMachine int
	*fslib.FsLib
	*procclnt.ProcClnt
}

func makeTstate(t *testing.T) *Tstate {
	setVersion()
	ts := &Tstate{}
	e := realm.MakeTestEnv(sp.TEST_RID)
	cfg, err := e.Boot()
	if err != nil {
		t.Fatalf("Boot %v\n", err)
	}
	ts.e = e
	ts.cfg = cfg

	err = ts.e.BootMachined()
	if err != nil {
		t.Fatalf("Boot Noded 2: %v", err)
	}

	program := "realm_test"
	ts.realmFsl, err = fslib.MakeFsLibAddr(program, ts.NamedAddr())
	assert.Nil(t, err)
	ts.ConfigClnt = config.MakeConfigClnt(ts.realmFsl)
	ts.FsLib, err = fslib.MakeFsLibAddr(program, cfg.NamedAddrs)
	assert.Nil(t, err)
	ts.ProcClnt = procclnt.MakeProcClntInit(proc.GenPid(), ts.FsLib, program, cfg.NamedAddrs)

	ts.t = t
	ts.coreGroupsPerMachine = int(1.0 / sp.Conf.Machine.CORE_GROUP_FRACTION)

	return ts
}

func setVersion() {
	if version == "" || version == "none" || !flag.Parsed() {
		db.DFatalf("Version not set in test")
	}
	proc.Version = version
}

func (ts *Tstate) spawnSpinner(ncore proc.Tcore) proc.Tpid {
	pid := proc.GenPid()
	a := proc.MakeProcPid(pid, "user/spinner", []string{"name/"})
	a.SetNcore(ncore)
	err := ts.Spawn(a)
	if err != nil {
		db.DFatalf("Error Spawn in RealmBalanceBenchmark.spawnSpinner: %v", err)
	}
	go func() {
		ts.WaitExit(pid)
	}()
	return pid
}

// Check that the test realm has min <= nCoreGroups <= max core groups assigned to it
func (ts *Tstate) checkNCoreGroups(min int, max int) {
	db.DPrintf(db.TEST, "Checking num nodeds")
	cfg := realm.GetRealmConfig(ts.realmFsl, sp.TEST_RID)
	nCoreGroups := 0
	for _, nd := range cfg.NodedsActive {
		ndCfg := realm.MakeNodedConfig()
		ts.ReadConfig(realm.NodedConfPath(nd), ndCfg)
		nCoreGroups += len(ndCfg.Cores)
	}
	db.DPrintf(db.TEST, "Done Checking num nodeds")
	ok := assert.True(ts.t, nCoreGroups >= min && nCoreGroups <= max, "Wrong number of core groups (x=%v), expected %v <= x <= %v", nCoreGroups, min, max)
	if !ok {
		debug.PrintStack()
	}
}

func TestStartStop(t *testing.T) {
	ts := makeTstate(t)
	ts.checkNCoreGroups(1, ts.coreGroupsPerMachine)
	ts.e.Shutdown()
}

func TestRealmGrowArtificial(t *testing.T) {
	ts := makeTstate(t)
	rclnt, err := realm.MakeRealmClnt()
	assert.Nil(t, err)
	for realm.GetRealmConfig(rclnt.FsLib, sp.TEST_RID).NCores < proc.Tcore(linuxsched.NCores)*2 {
		rclnt.GrowRealm(sp.TEST_RID)
	}
	ts.checkNCoreGroups(ts.coreGroupsPerMachine*2, ts.coreGroupsPerMachine*2)
	ts.e.Shutdown()
}

// Start enough spinning lambdas to fill two Nodeds, check that the test
// realm's allocation expanded, then exit.
func TestRealmGrowAuto(t *testing.T) {
	ts := makeTstate(t)

	N := int(linuxsched.NCores) / 2

	db.DPrintf(db.TEST, "Starting %v spinning lambdas", N)
	pids := []proc.Tpid{}
	for i := 0; i < N; i++ {
		pids = append(pids, ts.spawnSpinner(0))
	}

	db.DPrintf(db.TEST, "Sleeping for a bit")
	time.Sleep(SLEEP_TIME_MS * time.Millisecond)

	db.DPrintf(db.TEST, "Starting %v more spinning lambdas", N)
	for i := 0; i < N; i++ {
		pids = append(pids, ts.spawnSpinner(0))
	}

	db.DPrintf(db.TEST, "Sleeping again")
	time.Sleep(SLEEP_TIME_MS * time.Millisecond)

	ts.checkNCoreGroups(ts.coreGroupsPerMachine, ts.coreGroupsPerMachine*2)

	ts.e.Shutdown()
}

// Start enough spinning lambdas to fill two Nodeds, check that the test
// realm's allocation expanded, evict the spinning lambdas, and check the
// realm's allocation shrank. Then spawn and evict a few more to make sure we
// can still spawn after shrinking.  Assumes other machines in the cluster have
// the same number of cores.
func TestRealmShrink(t *testing.T) {
	ts := makeTstate(t)

	N := int(linuxsched.NCores) / 2

	db.DPrintf(db.TEST, "Starting %v spinning lambdas", N)
	pids := []proc.Tpid{}
	for i := 0; i < N; i++ {
		pids = append(pids, ts.spawnSpinner(0))
	}

	db.DPrintf(db.TEST, "Sleeping for a bit")
	time.Sleep(SLEEP_TIME_MS * time.Millisecond)

	db.DPrintf(db.TEST, "Starting %v more spinning lambdas", N)
	for i := 0; i < N; i++ {
		pids = append(pids, ts.spawnSpinner(0))
	}

	db.DPrintf(db.TEST, "Sleeping again")
	time.Sleep(SLEEP_TIME_MS * time.Millisecond)

	ts.checkNCoreGroups(ts.coreGroupsPerMachine, ts.coreGroupsPerMachine*2)

	db.DPrintf(db.TEST, "Creating a new realm to contend with the old one")
	// Create another realm to contend with this one.
	ts.e.CreateRealm("2000")

	db.DPrintf(db.TEST, "Sleeping yet again")
	time.Sleep(SLEEP_TIME_MS * time.Millisecond)

	ts.checkNCoreGroups(ts.coreGroupsPerMachine, ts.coreGroupsPerMachine*2-1)

	db.DPrintf(db.TEST, "Destroying the new, contending realm")
	ts.e.DestroyRealm("2000")

	db.DPrintf(db.TEST, "Starting %v more spinning lambdas", N/2)
	for i := 0; i < int(N); i++ {
		pids = append(pids, ts.spawnSpinner(0))
	}

	db.DPrintf(db.TEST, "Sleeping yet again")
	time.Sleep(SLEEP_TIME_MS * time.Millisecond)

	ts.checkNCoreGroups(ts.coreGroupsPerMachine, ts.coreGroupsPerMachine*2)

	ts.e.Shutdown()
}
