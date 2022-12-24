package kernel

import (
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	db "sigmaos/debug"
	"sigmaos/fslib"
	"sigmaos/kproc"
	"sigmaos/linuxsched"
	"sigmaos/proc"
	"sigmaos/procclnt"
	"sigmaos/sessp"
	sp "sigmaos/sigmap"
)

const (
	NO_PID           = "no-realm"
	NO_REALM         = "no-realm"
	SLEEP_MS         = 200
	REPL_PORT_OFFSET = 100
	SUBSYSTEM_INFO   = "subsystem-info"
)

type System struct {
	sync.Mutex
	*fslib.FsLib
	*procclnt.ProcClnt
	realmId     string
	namedAddr   []string
	procdIp     string
	named       *Subsystem
	cores       *sessp.Tinterval
	fss3d       []*Subsystem
	fsuxd       []*Subsystem
	procd       []*Subsystem
	dbd         []*Subsystem
	crashedPids map[proc.Tpid]bool
}

func makeSystemBase(realmId string, namedAddr []string, cores *sessp.Tinterval) *System {
	s := &System{}
	s.realmId = realmId
	s.namedAddr = namedAddr
	s.cores = cores
	s.procd = []*Subsystem{}
	s.fsuxd = []*Subsystem{}
	s.fss3d = []*Subsystem{}
	s.dbd = []*Subsystem{}
	s.crashedPids = make(map[proc.Tpid]bool)
	return s
}

// The boot processes enters here
func Boot(pn string) (*System, error) {
	db.DPrintf(db.KERNEL, "Boot %s\n", pn)
	param, err := readParam(pn)
	if err != nil {
		return nil, err
	}
	return makeSystem(param, MakeSystemNamed)
}

func (s *System) ShutDown() error {
	db.DPrintf(db.KERNEL, "ShutDown\n")
	s.Shutdown()
	db.DPrintf(db.KERNEL, "ShutDown done\n")
	return nil
}

// Make system with just named. replicaId is used to index into the
// fslib.Named() slice and select an address for this named.
func MakeSystemNamed(uname, realmId string, replicaId int, cores *sessp.Tinterval) (*System, error) {
	s := makeSystemBase(realmId, fslib.Named(), cores)
	// replicaId needs to be 1-indexed for replication library.
	cmd, err := RunNamed(fslib.Named()[replicaId], len(fslib.Named()) > 1, replicaId+1, fslib.Named(), NO_REALM)
	if err != nil {
		return nil, err
	}
	// XXX It's a bit weird that we set program/pid here...
	proc.SetProgram(uname)
	proc.SetPid(proc.GenPid())
	s.named = makeSubsystemCmd(nil, nil, "", false, cmd)
	time.Sleep(SLEEP_MS * time.Millisecond)
	s.FsLib, err = fslib.MakeFsLibAddr(uname, fslib.Named())
	return s, err
}

// func (ts *Tstate) addNamedReplica(i int) {
// 	defer ts.wg.Done()
// 	r := kernel.MakeSystemNamed("test", sp.TEST_RID, i, sessp.MkInterval(0, uint64(linuxsched.NCores)))
// 	ts.Lock()
// 	defer ts.Unlock()
// 	ts.replicas = append(ts.replicas, r)
// }

// func (ts *Tstate) startReplicas() {
// 	ts.replicas = []*kernel.System{}
// 	// Start additional replicas
// 	for i := 0; i < len(fslib.Named())-1; i++ {
// 		// Must happen in a separate thread because MakeSystemNamed
// 		// will block until the replicas are able to process requests.
// 		go ts.addNamedReplica(i + 1)
// 	}
// }

func makeSystem(p *Param, mkSys func(string, string, int, *sessp.Tinterval) (*System, error)) (*System, error) {
	wg := sync.WaitGroup{}
	db.DPrintf(db.KERNEL, "param %v\n", p)
	wg.Add(len(fslib.Named()))

	var sys *System
	var err error
	go func() {
		defer wg.Done()
		// Must happen in a separate thread because mkSys will block until
		// enough replicas have started (if named is replicated).
		sys, err = mkSys(p.Uname, p.Realm, 0, sessp.MkInterval(0, uint64(linuxsched.NCores)))
	}()
	// startReplicas()
	wg.Wait()
	return sys, err
}

// Make a system with Named and other kernel services
func MakeSystemAll(uname, realmId string, replicaId int, cores *sessp.Tinterval) (*System, error) {
	s, err := MakeSystemNamed(uname, realmId, replicaId, cores)
	if err != nil {
		db.DPrintf(db.KERNEL, "MakeSystemNamed err %v\n", err)
		return nil, err
	}
	// XXX should this be GetPid?
	s.ProcClnt = procclnt.MakeProcClntInit(proc.GenPid(), s.FsLib, uname, s.namedAddr)
	err = s.Boot()
	if err != nil {
		db.DPrintf(db.KERNEL, "Start err %v\n", err)
		return nil, err
	}
	return s, nil
}

func MakeSystem(uname, realmId string, namedAddr []string, cores *sessp.Tinterval) (*System, error) {
	s := makeSystemBase(realmId, namedAddr, cores)
	fsl, err := fslib.MakeFsLibAddr(uname, namedAddr)
	if err != nil {
		return nil, err
	}
	s.FsLib = fsl
	s.ProcClnt = procclnt.MakeProcClntInit(proc.GenPid(), s.FsLib, uname, namedAddr)
	return s, nil
}

// Boot a "kernel" without named
func (s *System) Boot() error {
	// Procd must boot first, since other services are spawned as procs.
	if err := s.bootProcd(true); err != nil {
		return err
	}
	if err := s.BootFsUxd(); err != nil {
		return err
	}
	if err := s.BootFss3d(); err != nil {
		return err
	}
	if err := s.BootDbd(); err != nil {
		return err
	}
	return nil
}

func (s *System) BootSubsystem(binpath string, args []string, procdIp string, viaProcd bool, list *[]*Subsystem) error {
	s.Lock()
	defer s.Unlock()

	pid := proc.Tpid(path.Base(binpath) + "-" + proc.GenPid().String())
	p := proc.MakeProcPid(pid, binpath, args)
	ss := makeSubsystem(s.ProcClnt, p, procdIp, viaProcd)
	// Lock appending to list
	*list = append(*list, ss)
	return ss.Run(s.namedAddr)
}

func (s *System) BootProcd() error {
	return s.bootProcd(false)
}

// Boot a procd. If spawningSys is true, procd will wait for all kernel procs
// to be spawned before claiming any procs.
func (s *System) bootProcd(spawningSys bool) error {
	err := s.BootSubsystem("kernel/procd", []string{path.Join(s.realmId, "bin"), s.cores.Marshal(), strconv.FormatBool(spawningSys)}, "", false, &s.procd)
	if err != nil {
		return err
	}
	if s.procdIp == "" {
		s.procdIp = s.GetProcdIp()
	}
	return nil
}

func (s *System) BootFsUxd() error {
	return s.BootSubsystem("kernel/fsuxd", []string{path.Join(sp.UXROOT, s.realmId)}, s.procdIp, true, &s.fsuxd)
}

func (s *System) BootFss3d() error {
	return s.BootSubsystem("kernel/fss3d", []string{s.realmId}, s.procdIp, true, &s.fss3d)
}

func (s *System) BootDbd() error {
	var dbdaddr string
	dbdaddr = os.Getenv("SIGMADBADDR")
	// XXX don't pass dbd addr as an envvar, it's messy.
	if dbdaddr == "" {
		dbdaddr = "127.0.0.1:3306"
	}
	return s.BootSubsystem("kernel/dbd", []string{dbdaddr}, s.procdIp, true, &s.dbd)
	return nil
}

func (s *System) GetProcdIp() string {
	s.Lock()
	defer s.Unlock()

	if len(s.procd) != 1 {
		db.DFatalf("Error unexpexted num procds: %v", s.procd)
	}
	return GetSubsystemInfo(s.FsLib, sp.KPIDS, s.procd[0].p.Pid.String()).Ip
}

func (s *System) KillOne(srv string) error {
	s.Lock()
	defer s.Unlock()

	var err error
	var ss *Subsystem
	switch srv {
	case sp.PROCD:
		if len(s.procd) > 0 {
			ss = s.procd[0]
			s.procd = s.procd[1:]
		} else {
			db.DPrintf(db.ALWAYS, "Tried to kill procd, nothing to kill")
		}
	case sp.UX:
		if len(s.fsuxd) > 0 {
			ss = s.fsuxd[0]
			s.fsuxd = s.fsuxd[1:]
		} else {
			db.DPrintf(db.ALWAYS, "Tried to kill ux, nothing to kill")
		}
	default:
		db.DFatalf("Unkown server type in System.KillOne: %v", srv)
	}
	err = ss.Kill()
	if err == nil {
		ss.Wait()
		s.crashedPids[ss.p.Pid] = true
	} else {
		db.DFatalf("%v kill failed %v\n", srv, err)
	}
	return nil
}

func (s *System) Shutdown() {
	if s.ProcClnt != nil {
		cpids, err := s.GetChildren()
		if err != nil {
			db.DFatalf("GetChildren in System.Shutdown: %v", err)
		}
		db.DPrintf(db.KERNEL, "Shutdown children %v", cpids)
		for _, pid := range cpids {
			s.Evict(pid)
			db.DPrintf(db.KERNEL, "Evicted %v", pid)
			if _, ok := s.crashedPids[pid]; !ok {
				if status, err := s.WaitExit(pid); err != nil || !status.IsStatusEvicted() {
					db.DPrintf(db.ALWAYS, "shutdown error pid %v: %v %v", pid, status, err)
				}
			}
			db.DPrintf(db.KERNEL, "Done evicting %v", pid)
		}
	}
	// Make sure the procs actually exited
	for _, d := range s.fss3d {
		d.Wait()
	}
	for _, d := range s.fsuxd {
		d.Wait()
	}
	for _, d := range s.procd {
		d.Wait()
	}
	for _, d := range s.dbd {
		d.Wait()
	}
	if s.named != nil {
		// kill it so that test terminates
		s.named.Terminate()
		s.named.Wait()
	}
}

func makeNamedProc(addr string, replicate bool, id int, pe []string, realmId string) *proc.Proc {
	args := []string{addr, realmId}
	// If we're running replicated...
	if replicate {
		// Add an offset to the peers' port addresses.
		peers := []string{}
		for _, peer := range pe {
			peers = append(peers, addReplPortOffset(peer))
		}
		args = append(args, strconv.Itoa(id))
		args = append(args, strings.Join(peers, ","))
	}

	return proc.MakeProcPid(proc.Tpid("pid-"+strconv.Itoa(id)+proc.GenPid().String()), "kernel/named", args)
}

// Run a named (but not as a proc)
func RunNamed(addr string, replicate bool, id int, peers []string, realmId string) (*exec.Cmd, error) {
	p := makeNamedProc(addr, replicate, id, peers, realmId)
	cmd, err := kproc.RunKernelProc(p, fslib.Named(), false)
	if err != nil {
		db.DPrintf(db.ALWAYS, "Error running named: %v", err)
		return nil, err
	}
	time.Sleep(SLEEP_MS * time.Millisecond)
	return cmd, nil
}

// Run a named as a proc
func BootNamed(pclnt *procclnt.ProcClnt, addr string, replicate bool, id int, peers []string, realmId string) (*exec.Cmd, proc.Tpid, error) {
	p := makeNamedProc(addr, replicate, id, peers, realmId)
	cmd, err := pclnt.SpawnKernelProc(p, fslib.Named(), "", false)
	if err != nil {
		db.DFatalf("Error SpawnKernelProc BootNamed: %v", err)
		return nil, "", err
	}
	if err = pclnt.WaitStart(p.Pid); err != nil {
		db.DFatalf("Error WaitStart in BootNamed: %v", err)
		return nil, "", err
	}
	return cmd, p.Pid, nil
}

func addReplPortOffset(peerAddr string) string {
	// Compute replica address as peerAddr + REPL_PORT_OFFSET
	host, port, err := net.SplitHostPort(peerAddr)
	if err != nil {
		db.DFatalf("Error splitting host port: %v", err)
	}
	portI, err := strconv.Atoi(port)
	if err != nil {
		db.DFatalf("Error conv port: %v", err)
	}
	newPort := strconv.Itoa(portI + REPL_PORT_OFFSET)

	return host + ":" + newPort
}
