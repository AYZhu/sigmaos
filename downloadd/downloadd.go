package downloadd

import (
	"errors"
	"path"
	"path/filepath"
	db "sigmaos/debug"
	"sigmaos/downloadd/proto"
	"sigmaos/downloaddclnt"
	"sigmaos/fs"
	"sigmaos/proc"
	"sigmaos/serr"
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

func (downloadsrv *Downloadd) DownloadLib(ctx fs.CtxI, req proto.DownloadLibRequest, rep *proto.DownloadLibResponse) error {
	db.DPrintf(DEBUG_DOWNLOAD_SERVER, "==%v== Received Download Request: %v\n", downloadsrv.kernelId, req)
	out, err := downloadsrv.tryDownloadLibPath(sp.Trealm(req.GetRealm()), req.GetNamedPath(), req.GetCopyFolder())
	if err == nil {
		rep.TmpPath = out
	}
	return nil
}

func InitDownloadPath() error {
	return nil
}

func (downloadsrv *Downloadd) touchLibPath(realm sp.Trealm, path string) error {
	db.DPrintf(db.ALWAYS, "touchLibPath %s", path)
	err := downloadsrv.client.MkDir(path, 0777)

	if err == nil {
		return nil
	}

	if serr.IsErrCode(err, serr.TErrNotfound) {
		err = downloadsrv.touchLibPath(realm, filepath.Dir(path))
	}

	if err != nil {
		return err
	}

	err = downloadsrv.client.MkDir(path, 0777)
	db.DPrintf(db.ALWAYS, "touchLibPathCompleteWith %s", path)
	return err
}

// Try to download a proc at pn to local Ux dir. May fail if ux crashes.
func (downloadsrv *Downloadd) tryDownloadLibPath(realm sp.Trealm, file string, copyFolder bool) (string, error) {
	// start := time.Now()
	db.DPrintf(db.ALWAYS, "tryDownloadProcPath %s", file)
	fs := downloadsrv.client
	libPath := path.Join(sp.UXBIN, "user", "common", file)
	fsfs, err := fs.IsDir(libPath)
	cachePn := path.Join(sp.UXBIN, "user", "realms", realm.String(), file)

	if err != nil {
		db.DPrintf(db.ALWAYS, "error 1 %s", err.Error())
		return "", err
	}

	if fsfs {
		err = fs.MkDir(cachePn, 0777)
		if err != nil {
			if serr.IsErrCode(err, serr.TErrNotfound) { // walk
				err = downloadsrv.touchLibPath(realm, cachePn)
			}
			if serr.IsErrCode(err, serr.TErrExists) {
				return cachePn, nil
			}
		}
		if err != nil {
			db.DPrintf(db.ALWAYS, "error 2 %s", err.Error())
			return "", err
		}
		if copyFolder {
			err = fs.CopyDir(libPath, cachePn)
			if err != nil {
				db.DPrintf(db.ALWAYS, "error 3 %s", err.Error())
				return "", err
			}
		}
		return cachePn, nil
	}

	if copyFolder {
		return "", errors.New("copy folder called on a file")
	}

	err = downloadsrv.client.CopyFile(libPath, cachePn)

	if serr.IsErrCode(err, serr.TErrExists) {
		return cachePn, nil
	}

	if err != nil {
		db.DPrintf(db.ALWAYS, "error! %s", err.Error())
		return "", err
	}

	return cachePn, nil
}
