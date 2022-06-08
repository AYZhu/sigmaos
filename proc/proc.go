package proc

import (
	"fmt"
	"log"
	"path"
	"strings"

	np "ulambda/ninep"
)

type Tpid string
type Ttype uint32
type Tcore uint32

const (
	T_DEF Ttype = 0
	T_LC  Ttype = 1
	T_BE  Ttype = 2
)

const (
	C_DEF Tcore = 0
)

func (pid Tpid) String() string {
	return string(pid)
}

type Proc struct {
	Pid          Tpid     // SigmaOS PID
	ProcDir      string   // SigmaOS directory to store this proc's state
	ParentDir    string   // SigmaOS parent proc directory
	Program      string   // Program to run
	Dir          string   // Unix working directory for the process
	Args         []string // Args
	Env          []string // Environment variables
	Type         Ttype    // Type
	Ncore        Tcore    // Number of cores requested
	sharedTarget string   // Target of shared state
}

func MakeEmptyProc() *Proc {
	p := &Proc{}
	return p
}

func MakeProc(program string, args []string) *Proc {
	p := &Proc{}
	p.Pid = GenPid()
	p.Program = program
	p.Args = args
	p.Type = T_DEF
	p.Ncore = C_DEF
	p.setProcDir("")
	// If this isn't a user proc, version it.
	if !p.IsPrivilegedProc() {
		// Check the version has been set.
		if Version == "none" {
			log.Fatalf("FATAL %v %v Version not set. Please set by running with --version", GetName(), GetPid())
		}
		// Set the Program to user/VERSION/prog.bin
		p.Program = path.Join(path.Dir(p.Program), Version, path.Base(p.Program))
	}
	return p
}

func MakeProcPid(pid Tpid, program string, args []string) *Proc {
	p := MakeProc(program, args)
	p.Pid = pid
	p.setProcDir("")
	return p
}

// Called by procclnt to set the parent dir when spawning.
func (p *Proc) SetParentDir(parentdir string) {
	if parentdir == PROCDIR {
		p.ParentDir = path.Join(GetProcDir(), CHILDREN, p.Pid.String())
	} else {
		p.ParentDir = path.Join(parentdir, CHILDREN, p.Pid.String())
	}
}

func (p *Proc) setProcDir(procdIp string) {
	if p.IsPrivilegedProc() {
		p.ProcDir = path.Join(KPIDS, p.Pid.String())
	} else {
		if procdIp != "" {
			p.ProcDir = path.Join(np.PROCD, procdIp, PIDS, p.Pid.String()) // TODO: make relative to ~procd
		}
	}
}

func (p *Proc) AppendEnv(name, val string) {
	p.Env = append(p.Env, name+"="+val)
}

func (p *Proc) GetEnv(procdIp, newRoot string) []string {
	// Set the procdir based on procdIp
	p.setProcDir(procdIp)

	env := []string{}
	for _, envvar := range p.Env {
		env = append(env, envvar)
	}
	env = append(env, SIGMAPRIVILEGEDPROC+"="+fmt.Sprintf("%v", p.IsPrivilegedProc()))
	env = append(env, SIGMANEWROOT+"="+newRoot)
	env = append(env, SIGMAPROCDIP+"="+procdIp)
	env = append(env, SIGMAPID+"="+p.Pid.String())
	env = append(env, SIGMAPROGRAM+"="+p.Program)
	env = append(env, SIGMAPROCDIR+"="+p.ProcDir)
	env = append(env, SIGMAPARENTDIR+"="+p.ParentDir)
	env = append(env, SIGMANODEDID+"="+GetNodedId())
	return env
}

func (p *Proc) SetShared(target string) {
	p.sharedTarget = target
}

func (p *Proc) GetShared() string {
	return p.sharedTarget
}

func (p *Proc) IsPrivilegedProc() bool {
	return strings.Contains(p.Program, "kernel") || strings.Contains(p.Program, "realm")
}

func (p *Proc) String() string {
	return fmt.Sprintf("&{ Pid:%v Program:%v ProcDir:%v ParentDir:%v UnixDir:%v Args:%v Env:%v Type:%v Ncore:%v }", p.Pid, p.Program, p.ProcDir, p.ParentDir, p.Dir, p.Args, p.GetEnv("", ""), p.Type, p.Ncore)
}
