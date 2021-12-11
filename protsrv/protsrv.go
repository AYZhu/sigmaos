package protsrv

import (
	np "ulambda/ninep"
)

type FsServer interface {
	Dispatch(sess np.Tsession, msg np.Tmsg) (np.Tmsg, *np.Rerror)
	Detach(np.Tsession)
}

type Protsrv interface {
	Version(np.Tversion, *np.Rversion) *np.Rerror
	Auth(np.Tauth, *np.Rauth) *np.Rerror
	Flush(np.Tflush, *np.Rflush) *np.Rerror
	Attach(np.Tattach, *np.Rattach) *np.Rerror
	Walk(np.Twalk, *np.Rwalk) *np.Rerror
	Create(np.Tcreate, *np.Rcreate) *np.Rerror
	Open(np.Topen, *np.Ropen) *np.Rerror
	WatchV(np.Twatchv, *np.Ropen) *np.Rerror
	Clunk(np.Tclunk, *np.Rclunk) *np.Rerror
	Read(np.Tread, *np.Rread) *np.Rerror
	Write(np.Twrite, *np.Rwrite) *np.Rerror
	Remove(np.Tremove, *np.Rremove) *np.Rerror
	RemoveFile(np.Tremovefile, *np.Rremove) *np.Rerror
	Stat(np.Tstat, *np.Rstat) *np.Rerror
	Wstat(np.Twstat, *np.Rwstat) *np.Rerror
	Renameat(np.Trenameat, *np.Rrenameat) *np.Rerror
	GetFile(np.Tgetfile, *np.Rgetfile) *np.Rerror
	SetFile(np.Tsetfile, *np.Rwrite) *np.Rerror
	Detach()
}

type MkProtServer func(FsServer, np.Tsession) Protsrv
