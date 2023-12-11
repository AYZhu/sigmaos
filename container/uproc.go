package container

import (
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
	"time"

	db "sigmaos/debug"
	"sigmaos/linuxsched"
	"sigmaos/proc"
	sp "sigmaos/sigmap"
)

//
// Contain user procs using exec-uproc-rs trampoline
//py

func printDir(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		db.DPrintf(db.TEST, "entry %v/%v\n", dir, file.Name())
	}
}

func RunUProc(uproc *proc.Proc) error {
	db.DPrintf(db.CONTAINER, "RunUProc %v env %v\n", uproc, os.Environ())

	printDir("/home/sigmaos/bin/user")

	// cmd := exec.Command("strace", append([]string{"-f", "exec-uproc-rs", uproc.GetProgram()}, uproc.Args...)...)
	cmd := exec.Command("exec-uproc-rs", append([]string{uproc.GetProgram()}, uproc.Args...)...)
	uproc.AppendEnv("PATH", "/bin:/bin2:/usr/bin:/home/sigmaos/bin/kernel")
	uproc.AppendEnv("SIGMA_EXEC_TIME", strconv.FormatInt(time.Now().UnixMicro(), 10))
	uproc.AppendEnv("RUST_BACKTRACE", "1")
	uproc.AppendEnv("PYTHONPATH", "/~~/pylib/Lib") // todo: fix
	uproc.AppendEnv("LD_PRELOAD", "/bin2/ld_fstatat.so")
	cmd.Env = uproc.GetEnv()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Set up new namespaces
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS,
	}
	db.DPrintf(db.CONTAINER, "exec %v\n", cmd)
	defer cleanupJail(uproc.GetPid())
	s := time.Now()
	if err := cmd.Start(); err != nil {
		db.DPrintf(db.CONTAINER, "Error start %v %v", cmd, err)
		return err
	}
	db.DPrintf(db.SPAWN_LAT, "[%v] Uproc cmd.Start %v", uproc.GetPid(), time.Since(s))
	if uproc.GetType() == proc.T_BE {
		s := time.Now()
		setSchedPolicy(cmd.Process.Pid, linuxsched.SCHED_IDLE)
		db.DPrintf(db.SPAWN_LAT, "[%v] Uproc Get/Set sched attr %v", uproc.GetPid(), time.Since(s))
	}

	if err := cmd.Wait(); err != nil {
		return err
	}
	db.DPrintf(db.CONTAINER, "ExecUProc done  %v\n", uproc)
	return nil
}

// Clean up a proc's chroot jail.
func cleanupJail(pid sp.Tpid) {
	if err := os.RemoveAll(jailPath(pid)); err != nil {
		db.DPrintf(db.ALWAYS, "Error cleanupJail: %v", err)
	}
}

func setSchedPolicy(pid int, policy linuxsched.SchedPolicy) {
	attr, err := linuxsched.SchedGetAttr(pid)
	if err != nil {
		db.DPrintf(db.ALWAYS, "Error Getattr %v: %v", pid, err)
		return
	}
	attr.Policy = policy
	err = linuxsched.SchedSetAttr(pid, attr)
	if err != nil {
		db.DPrintf(db.ALWAYS, "Error Setattr %v: %v", pid, err)
	}
}

func jailPath(pid sp.Tpid) string {
	return path.Join(sp.SIGMAHOME, "jail", pid.String())
}
