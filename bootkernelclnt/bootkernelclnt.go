package bootkernelclnt

import (
	"os/exec"
	"path"

	db "sigmaos/debug"
	"sigmaos/kernelclnt"
	"sigmaos/rand"
	"sigmaos/sigmaclnt"
	sp "sigmaos/sigmap"
)

//
// Library to start kernel
//

const (
	START = "../start-kernel.sh"
)

func Start(kernelId, tag, srvs string, namedAddr sp.Taddrs, overlays bool) (string, error) {
	s, e := namedAddr.Taddrs2String()
	if e != nil {
		return "", e
	}
	args := []string{
		"--pull", tag,
		"--boot", srvs,
		"--named", s,
		"--host",
	}
	if overlays {
		args = append(args, "--overlays")
	}
	args = append(args, kernelId)
	out, err := exec.Command(START, args...).Output()
	if err != nil {
		db.DPrintf(db.BOOT, "Boot: start out %s err %v\n", string(out), err)
		return "", err
	}
	ip := string(out)
	db.DPrintf(db.BOOT, "Start: %v srvs %v IP %v\n", kernelId, srvs, ip)
	return ip, nil
}

func GenKernelId() string {
	return "sigma-" + rand.String(4)
}

type Kernel struct {
	*sigmaclnt.SigmaClnt
	kernelId string
	kclnt    *kernelclnt.KernelClnt
}

func MkKernelClntStart(tag string, uname sp.Tuname, conf string, namedAddr sp.Taddrs, overlays bool) (*Kernel, error) {
	kernelId := GenKernelId()
	ip, err := Start(kernelId, tag, conf, namedAddr, overlays)
	if err != nil {
		return nil, err
	}
	return MkKernelClnt(kernelId, uname, ip, namedAddr)
}

func MkKernelClnt(kernelId string, uname sp.Tuname, ip string, namedAddr sp.Taddrs) (*Kernel, error) {
	db.DPrintf(db.SYSTEM, "MakeKernelClnt %s\n", kernelId)
	sc, err := sigmaclnt.MkSigmaClntRootInit(uname, ip, namedAddr)
	if err != nil {
		db.DPrintf(db.ALWAYS, "Error make sigma clnt root init")
		return nil, err
	}
	pn := sp.BOOT + kernelId
	if kernelId == "" {
		pn1, _, err := sc.ResolveUnion(sp.BOOT + "~local")
		if err != nil {
			db.DPrintf(db.ALWAYS, "Error resolve local")
			return nil, err
		}
		pn = pn1
		kernelId = path.Base(pn)
	}

	db.DPrintf(db.SYSTEM, "MakeKernelClnt %s %s\n", pn, kernelId)
	kclnt, err := kernelclnt.MakeKernelClnt(sc.FsLib, pn)
	if err != nil {
		db.DPrintf(db.ALWAYS, "Error MkKernelClnt")
		return nil, err
	}

	return &Kernel{sc, kernelId, kclnt}, nil
}

func (k *Kernel) MkSigmaClnt(uname sp.Tuname) (*sigmaclnt.SigmaClnt, error) {
	return sigmaclnt.MkSigmaClntRootInit(uname, k.GetLocalIP(), k.SigmaClnt.NamedAddr())
}

func (k *Kernel) Shutdown() error {
	db.DPrintf(db.SYSTEM, "Shutdown kernel %s", k.kernelId)
	err := k.kclnt.Shutdown()
	db.DPrintf(db.SYSTEM, "Shutdown kernel %s err %v", k.kernelId, err)
	return err
}

func (k *Kernel) Boot(s string) error {
	_, err := k.kclnt.Boot(s, []string{})
	return err
}

func (k *Kernel) Kill(s string) error {
	return k.kclnt.Kill(s)
}

func (k *Kernel) KernelId() string {
	return k.kernelId
}
