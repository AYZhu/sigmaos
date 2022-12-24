package bootclnt

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"syscall"

	"sigmaos/container"
	db "sigmaos/debug"
	sp "sigmaos/sigmap"
)

//
// Library to start a kernel boot process
//

type Kernel struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func BootKernel(contain bool) (*Kernel, error) {
	pn := path.Join(sp.PRIVILEGED_BIN, "kernel")
	cmd := exec.Command(pn+"/boot", []string{pn + "/boot.yml"}...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = os.Stderr

	if contain {
		if err := container.RunContainer(cmd); err != nil {
			return nil, err
		}
	} else {
		// Create a process group ID to kill all children if necessary.
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if err := cmd.Start(); err != nil {
			db.DPrintf(db.KERNEL, "BootKernel: Start err %v\n", err)
			return nil, err
		}
	}

	db.DPrintf(db.KERNEL, "Wait for kernel to be booted\n")
	// wait for kernel to be booted
	s := ""
	if _, err := fmt.Fscanf(stdout, "%s", &s); err != nil {
		db.DPrintf(db.KERNEL, "Fscanf err %v %s\n", err, s)
		return nil, err
	}
	db.DPrintf(db.KERNEL, "Kernel is running\n")
	return &Kernel{cmd, stdin, stdout}, nil
}

func (k *Kernel) Shutdown() error {
	defer k.stdout.Close()
	if _, err := io.WriteString(k.stdin, "shutdown\n"); err != nil {
		return err
	}
	defer k.stdin.Close()
	db.DPrintf(db.KERNEL, "Wait for kernel to shutdown\n")
	if err := k.cmd.Wait(); err != nil {
		return err
	}
	container.DelScnet(k.cmd.Process.Pid)
	return nil
}
