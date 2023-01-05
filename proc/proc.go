package proc

import (
	"fmt"
	"log"
	"path"
	"time"

	sp "sigmaos/sigmap"
)

type Tpid string
type Ttype uint32
type Tcore uint32
type Tmem uint32

const (
	T_BE Ttype = 0
	T_LC Ttype = 1
)

const (
	C_DEF Tcore = 0
)

func (t Ttype) String() string {
	switch t {
	case T_BE:
		return "T_BE"
	case T_LC:
		return "T_LC"
	default:
		log.Fatalf("Unknown proc type: %v", t)
	}
	return ""
}

func (pid Tpid) String() string {
	return string(pid)
}

type Proc struct {
	Pid          Tpid      // SigmaOS PID
	Privileged   bool      // kernel proc?
	ProcDir      string    // SigmaOS directory to store this proc's state
	ParentDir    string    // SigmaOS parent proc directory
	Program      string    // Program to run
	Args         []string  // Args
	Env          []string  // Environment variables
	Type         Ttype     // Type
	Ncore        Tcore     // Number of cores requested
	Mem          Tmem      // Amount of memory required in MB
	SpawnTime    time.Time // Time at which the proc was spawned
	sharedTarget string    // Target of shared state
}

func MakeEmptyProc() *Proc {
	p := &Proc{}
	return p
}

func MakeProc(program string, args []string) *Proc {
	pid := GenPid()
	return MakeProcPid(pid, program, args)
}

func MakePrivProcPid(pid Tpid, program string, args []string, priv bool) *Proc {
	p := &Proc{}
	p.Pid = pid
	p.Program = program
	p.Args = args
	p.Type = T_BE
	p.Ncore = C_DEF
	p.Privileged = priv
	p.setProcDir("")
	if !p.Privileged {
		p.Type = T_LC
	}
	p.setBaseEnv()
	return p
}

func MakeProcPid(pid Tpid, program string, args []string) *Proc {
	return MakePrivProcPid(pid, program, args, false)
}

// Called by procclnt to set the parent dir when spawning.
func (p *Proc) SetParentDir(parentdir string) {
	if parentdir == PROCDIR {
		p.ParentDir = path.Join(GetProcDir(), CHILDREN, p.Pid.String())
	} else {
		p.ParentDir = path.Join(parentdir, CHILDREN, p.Pid.String())
	}
}

// Set the number of cores on this proc. If > 0, then this proc is LC. For now,
// LC procs necessarily must specify LC > 1.
func (p *Proc) SetNcore(ncore Tcore) {
	if ncore > Tcore(0) {
		p.Type = T_LC
		p.Ncore = ncore
	}
}

// Set the amount of memory (in MB) required to run this proc.
func (p *Proc) SetMem(mb Tmem) {
	p.Mem = mb
}

func (p *Proc) setProcDir(procdIp string) {
	if p.IsPrivilegedProc() {
		p.ProcDir = path.Join(KPIDS, p.Pid.String())
	} else {
		if procdIp != "" {
			p.ProcDir = path.Join(sp.PROCD, procdIp, PIDS, p.Pid.String())
		}
	}
}

func (p *Proc) AppendEnv(name, val string) {
	p.Env = append(p.Env, name+"="+val)
}

// Set the envvars which can be set at proc creation time.
func (p *Proc) setBaseEnv() {
	p.AppendEnv(SIGMAPRIVILEGEDPROC, fmt.Sprintf("%t", p.IsPrivilegedProc()))
	p.AppendEnv(SIGMAPID, p.Pid.String())
	p.AppendEnv(SIGMAPROGRAM, p.Program)
	// Pass through debug/performance vars.
	p.AppendEnv(SIGMAPERF, GetSigmaPerf())
	p.AppendEnv(SIGMADEBUG, GetSigmaDebug())
	p.AppendEnv(SIGMANAMED, GetSigmaNamed())
	if p.Privileged {
		p.AppendEnv(PATH, GetPath()) // inherit from boot
	}
}

// Finalize env details which can only be set once a physical machine has been
// chosen.
func (p *Proc) FinalizeEnv(procdIp string) {
	// Set the procdir based on procdIp
	p.setProcDir(procdIp)
	p.AppendEnv(SIGMAPROCDIP, procdIp)
	p.AppendEnv(SIGMANODEDID, GetNodedId())
	p.AppendEnv(SIGMAPROCDIR, p.ProcDir)
	p.AppendEnv(SIGMAPARENTDIR, p.ParentDir)
}

func (p *Proc) GetEnv() []string {
	env := []string{}
	for _, envvar := range p.Env {
		env = append(env, envvar)
	}
	return env
}

func (p *Proc) SetShared(target string) {
	p.sharedTarget = target
}

func (p *Proc) GetShared() string {
	return p.sharedTarget
}

func (p *Proc) IsPrivilegedProc() bool {
	return p.Privileged
}

func (p *Proc) String() string {
	return fmt.Sprintf("&{ Pid:%v Priv %t Program:%v ProcDir:%v ParentDir:%v UnixDir:%v Args:%v Env:%v Type:%v Ncore:%v }", p.Pid, p.Privileged, p.Program, p.ProcDir, p.ParentDir, "Abcd", p.Args, p.GetEnv(), p.Type, p.Ncore)
}
