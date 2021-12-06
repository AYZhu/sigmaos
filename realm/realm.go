package realm

import (
	"fmt"
	"log"
	"path"
	"time"

	"ulambda/config"
	"ulambda/fslib"
	"ulambda/named"
	"ulambda/procclnt"
	"ulambda/sync"
)

const (
	DEFAULT_REALM_PRIORITY = "0"
	MIN_PORT               = 1112
	MAX_PORT               = 60000
)

type RealmConfig struct {
	Rid        string    // Realm id.
	NMachineds int       // Number of machineds currently assigned to this realm.
	LastResize time.Time // Timestamp from the last time this realm was resized
	Shutdown   bool      // True if this realm is in the process of being destroyed.
	NamedAddr  []string  // IP address of this realm's nameds.
	NamedPids  []string  // PIDs of this realm's nameds.
}

type RealmClnt struct {
	*procclnt.ProcClnt
	*fslib.FsLib
}

func MakeRealmClnt() *RealmClnt {
	clnt := &RealmClnt{}
	clnt.FsLib = fslib.MakeFsLib(fmt.Sprintf("realm-clnt"))
	clnt.ProcClnt = procclnt.MakeProcClntInit(clnt.FsLib, fslib.Named())
	return clnt
}

// Submit a realm creation request to the realm manager, and wait for the
// request to be handled.
func (clnt *RealmClnt) CreateRealm(rid string) *RealmConfig {
	// Create cond var to wait on realm creation/initialization.
	rStartCond := sync.MakeCond(clnt.FsLib, path.Join(named.BOOT, rid), nil, true)
	rStartCond.Init()

	if err := clnt.WriteFile(REALM_CREATE, []byte(rid)); err != nil {
		log.Fatalf("Error WriteFile in RealmClnt.CreateRealm: %v", err)
	}

	// Wait for the realm to be initialized
	rStartCond.Wait()

	return GetRealmConfig(clnt.FsLib, rid)
}

func (clnt *RealmClnt) DestroyRealm(rid string) {
	// Create cond var to wait on realm creation/initialization.
	rExitCond := sync.MakeCond(clnt.FsLib, path.Join(named.BOOT, rid), nil, true)
	rExitCond.Init()

	if err := clnt.WriteFile(REALM_DESTROY, []byte(rid)); err != nil {
		log.Fatalf("Error WriteFile in RealmClnt.DestroyRealm: %v", err)
	}

	rExitCond.Wait()
}

// Get a realm's configuration
func GetRealmConfig(fsl *fslib.FsLib, rid string) *RealmConfig {
	clnt := config.MakeConfigClnt(fsl)
	cfg := &RealmConfig{}
	clnt.ReadConfig(path.Join(REALM_CONFIG, rid), cfg)
	return cfg
}
