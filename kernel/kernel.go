package kernel

import (
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"sigmaos/container"
	db "sigmaos/debug"
	"sigmaos/fslib"
	"sigmaos/kproc"
	"sigmaos/linuxsched"
	"sigmaos/proc"
	"sigmaos/procclnt"
	"sigmaos/sessp"
	"sigmaos/sigmaclnt"
	sp "sigmaos/sigmap"
)

const (
	NO_PID           = "no-realm"
	NO_REALM         = "norealm"
	SLEEP_MS         = 200
	REPL_PORT_OFFSET = 100
	SUBSYSTEM_INFO   = "subsystem-info"
)

type Param struct {
	Realm    string   `yalm:"realm, omitempty"`
	Hostip   string   `yalm:"ip, omitempty"`
	Services []string `yalm:"services"`
}

type Kernel struct {
	*sigmaclnt.SigmaClnt
	Param     *Param
	namedAddr sp.Taddrs
	procdIp   string
	cores     *sessp.Tinterval
	svcs      *Services
	ip        string
}

func mkKernel(param *Param, namedAddr sp.Taddrs, cores *sessp.Tinterval) *Kernel {
	k := &Kernel{}
	k.Param = param
	k.namedAddr = namedAddr
	k.cores = cores
	k.svcs = mkServices()
	return k
}

func MakeKernel(p *Param, nameds sp.Taddrs) (*Kernel, error) {
	cores := sessp.MkInterval(0, uint64(linuxsched.NCores))
	k := mkKernel(p, nameds, cores)
	proc.SetProgram(os.Args[0])
	proc.SetPid(proc.GenPid())
	ip, err := container.LocalIP()
	if err != nil {
		return nil, err
	}
	db.DPrintf(db.KERNEL, "MakeKernel ip %v", ip)
	k.ip = ip
	proc.SetSigmaLocal(ip)
	if p.Services[0] == sp.NAMEDREL {
		k.makeNameds()
		nameds, err := SetNamedIP(k.ip, k.namedAddr)
		if err != nil {
			return nil, err
		}
		k.namedAddr = nameds
		p.Services = p.Services[1:]
	}
	proc.SetSigmaNamed(k.namedAddr)
	sc, err := sigmaclnt.MkSigmaClntProc("kernel", ip, k.namedAddr)
	if err != nil {
		db.DPrintf(db.ALWAYS, "Error MkSigmaClntProc (%v): %v", k.namedAddr, err)
		return nil, err
	}
	k.SigmaClnt = sc
	err = startSrvs(k)
	if err != nil {
		db.DPrintf(db.ALWAYS, "Error startSrvs %v", err)
		return nil, err
	}
	return k, err
}

func (k *Kernel) Ip() string {
	return k.ip
}

func (k *Kernel) Shutdown() error {
	db.DPrintf(db.KERNEL, "ShutDown\n")
	k.shutdown()
	N := 200 // Crashing procds in mr test leave several fids open; maybe too many?
	n := k.PathClnt.FidClnt.Len()
	if n > N {
		log.Printf("Too many FIDs open (%v): %v", n, k.PathClnt.FidClnt)
	}
	db.DPrintf(db.KERNEL, "ShutDown done\n")
	return nil
}

// Start nameds and wait until they have started
func (k *Kernel) makeNameds() error {
	n := len(k.namedAddr)
	ch := make(chan error)
	k.startNameds(ch, n)
	var err error
	for i := 0; i < n; i++ {
		r := <-ch
		if r != nil {
			err = r
		}
	}
	return err
}

func (k *Kernel) startNameds(ch chan error, n int) {
	for i := 0; i < n; i++ {
		// Must happen in a separate thread because MakeKernelNamed
		// will block until the replicas are able to process requests.
		go func(i int) {
			err := bootNamed(k, k.Param.Realm, i, k.Param.Realm)
			ch <- err
		}(i)
	}
}

// Start kernel services listed in p
func startSrvs(k *Kernel) error {
	n := len(k.Param.Services)
	for _, s := range k.Param.Services {
		err := k.BootSub(s, nil, k.Param, n > 1) // XXX kernel should wait instead of procd?
		if err != nil {
			db.DPrintf(db.KERNEL, "Start %s err %v\n", k.Param, err)
			return err
		}
	}
	return nil
}

func (k *Kernel) shutdown() {
	if len(k.Param.Services) > 0 {
		db.DPrintf(db.KERNEL, "Get children %v", proc.GetPid())
		cpids, err := k.GetChildren()
		if err != nil {
			db.DPrintf(db.KERNEL, "Error get children: %v", err)
			db.DFatalf("GetChildren in Kernel.Shutdown: %v", err)
		}
		db.DPrintf(db.KERNEL, "Shutdown children %v", cpids)
		for _, pid := range cpids {
			k.Evict(pid)
			db.DPrintf(db.KERNEL, "Evicted %v", pid)
			if _, ok := k.svcs.crashedPids[pid]; !ok {
				if status, err := k.WaitExit(pid); err != nil || !status.IsStatusEvicted() {
					db.DPrintf(db.ALWAYS, "shutdown error pid %v: %v %v", pid, status, err)
				}
			}
			db.DPrintf(db.KERNEL, "Done evicting %v", pid)
		}
	}
	for key, val := range k.svcs.svcs {
		if key != sp.NAMEDREL {
			for _, d := range val {
				d.Wait()
			}

		}
	}
	for _, d := range k.svcs.svcs[sp.NAMEDREL] {
		// kill it so that test terminates
		d.Terminate()
		d.Wait()
	}
}

func makeNamedProc(addr string, replicate bool, id int, pe sp.Taddrs, realmId string) *proc.Proc {
	args := []string{addr, realmId, ""}
	// If we're running replicated...
	if replicate {
		// Add an offset to the peers' port addresses.
		peers := sp.Taddrs{}
		for _, peer := range pe {
			peers = append(peers, addReplPortOffset(peer))
		}
		args = append(args, strconv.Itoa(id))
		args = append(args, strings.Join(peers, ","))
	}

	p := proc.MakePrivProcPid(proc.Tpid("pid-"+strconv.Itoa(id)+proc.GenPid().String()), "named", args, true)
	return p
}

// Run a named (but not as a proc)
func RunNamed(addr string, replicate bool, id int, peers []string, realmId string) (*exec.Cmd, error) {
	p := makeNamedProc(addr, replicate, id, peers, realmId)
	cmd, err := kproc.RunKernelProc(p, peers, realmId)
	if err != nil {
		db.DPrintf(db.ALWAYS, "Error running named: %v", err)
		return nil, err
	}
	time.Sleep(SLEEP_MS * time.Millisecond)
	return cmd, nil
}

func SetNamedIP(ip string, ports []string) ([]string, error) {
	nameds := make([]string, len(ports))
	for i, s := range ports {
		host, port, err := net.SplitHostPort(s)
		if err != nil {
			return nil, err
		}
		if host != "" {
			db.DFatalf("Tried to substitute named ip when port exists: %v -> %v %v", s, host, port)
		}
		nameds[i] = net.JoinHostPort(ip, port)
	}
	return nameds, nil
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

//
// XXX kill backward-compatability, but keep for now for noded.go.
//

func MakeSystem(uname, realmId string, namedAddr sp.Taddrs, cores *sessp.Tinterval) (*Kernel, error) {
	p := &Param{Realm: realmId}
	s := mkKernel(p, namedAddr, cores)
	fsl, err := fslib.MakeFsLibAddr(p.Realm, s.ip, namedAddr)
	if err != nil {
		return nil, err
	}
	s.FsLib = fsl
	s.ProcClnt = procclnt.MakeProcClntInit(proc.GenPid(), s.FsLib, p.Realm, namedAddr)
	return s, nil
}

// Run a named as a proc
func BootNamed(pclnt *procclnt.ProcClnt, addr string, replicate bool, id int, peers sp.Taddrs, realmId string) (*exec.Cmd, proc.Tpid, error) {
	p := makeNamedProc(addr, replicate, id, peers, realmId)
	cmd, err := pclnt.SpawnKernelProc(p, pclnt.NamedAddr(), realmId, procclnt.HLINUX)
	if err != nil {
		db.DFatalf("Error SpawnKernelProc BootNamed: %v", err)
		return nil, "", err
	}
	if err = pclnt.WaitStart(p.GetPid()); err != nil {
		db.DFatalf("Error WaitStart in BootNamed: %v", err)
		return nil, "", err
	}
	return cmd, p.GetPid(), nil
}

// Boot subsystems other than named
func (k *Kernel) BootSubs() error {
	// Procd must boot first, since other services are spawned as
	// procs.
	for _, s := range []string{sp.PROCDREL, sp.S3REL, sp.UXREL, sp.DBREL} {
		err := k.BootSub(s, nil, nil, true)
		if err != nil {
			return err
		}
	}
	return nil
}
