package downloadd

import (
	"path"
	db "sigmaos/debug"
	"sigmaos/downloadd/proto"
	"sigmaos/downloaddclnt"
	"sigmaos/fs"
	"sigmaos/proc"
	"sigmaos/sigmaclnt"
	"sigmaos/sigmasrv"

	sp "sigmaos/sigmap"
)

type Downloadd struct {
	kernelId string
	realms   []sp.Trealm
	client   sigmaclnt.SigmaClnt
	ddc      *downloaddclnt.DownloaddClnt
}

const DEBUG_DOWNLOAD_SERVER = "DOWNLOAD_SERVER"
const DIR_DOWNLOAD_SERVER = sp.NAMED + "downloadd/"

func NewDownloadd(kernelId string) *Downloadd {
	sd := &Downloadd{
		realms:   make([]sp.Trealm, 0),
		kernelId: kernelId,
	}
	return sd
}

func RunDownloadd(kernelId string) error {
	sd := NewDownloadd(kernelId)
	ssrv, err := sigmasrv.NewSigmaSrv(path.Join(DIR_DOWNLOAD_SERVER, kernelId), sd, proc.GetProcEnv())
	sd.client = *ssrv.MemFs.SigmaClnt()
	// sd.ddc, err = downloaddclnt.NewDownloaddClnt(sd.client.FsLib, kernelId)
	if err != nil {
		db.DFatalf("Error PDS: %v", err)
	}
	ssrv.RunServer()
	return nil
}

func (downloadsrv *Downloadd) Download(ctx fs.CtxI, req proto.DownloadRequest, rep *proto.DownloadResponse) error {
	db.DPrintf(DEBUG_DOWNLOAD_SERVER, "==%v== Received Download Request: %v\n", downloadsrv.kernelId, req)
	out, err := downloadsrv.tryDownloadPath("<TODO>", req.NamedPath) // NamedPath should probably be path.Join(sp.UXBIN, "user", "common", "filename")
	if err == nil {
		rep.TmpPath = out
	}
	return nil
}

func InitDownloadPath() error {
	return nil
}

// Try to download a proc at pn to local Ux dir. May fail if ux crashes.
func (downloadsrv *Downloadd) tryDownloadPath(realm sp.Trealm, file string) (string, error) {
	// start := time.Now()
	// db.DPrintf(db.PROCMGR, "tryDownloadProcPath %s", src)
	// cachePn := path.Join(sp.UX, "lib", "user", "realms", realm.String(), file) TODO enable
	cachePn := path.Join(sp.UXBIN, "user", "realms", realm.String(), file)
	// Copy the binary from s3 to a temporary file.
	if err := downloadsrv.client.CopyFile(file, cachePn); err != nil {
		return "", err
	}
	// db.DPrintf(db.PROCMGR, "Took %v to download proc %v", time.Since(start), src)
	return cachePn, nil
}
