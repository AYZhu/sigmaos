package sigmap

import (
	"sigmaos/fcall"
)

type Isrvconn interface {
}

type Conn interface {
	IsClosed() bool
	Close()
	CloseConnTest()
	GetReplyC() chan *FcallMsg
}

type Fsrvfcall func(fcall.Fcall)

type SessServer interface {
	Register(fcall.Tclient, fcall.Tsession, Conn) *fcall.Err
	Unregister(fcall.Tclient, fcall.Tsession, Conn)
	SrvFcall(fcall.Fcall)
	Snapshot() []byte
	Restore([]byte)
}

type Protsrv interface {
	Version(*Tversion, *Rversion) *Rerror
	Auth(*Tauth, *Rauth) *Rerror
	Flush(*Tflush, *Rflush) *Rerror
	Attach(*Tattach, *Rattach) *Rerror
	Walk(*Twalk, *Rwalk) *Rerror
	Create(*Tcreate, *Rcreate) *Rerror
	Open(*Topen, *Ropen) *Rerror
	Watch(*Twatch, *Ropen) *Rerror
	Clunk(*Tclunk, *Rclunk) *Rerror
	Read(*Tread, *Rread) *Rerror
	ReadV(*TreadV, *Rread) *Rerror
	Write(*Twrite, *Rwrite) *Rerror
	WriteV(*TwriteV, *Rwrite) *Rerror
	Remove(*Tremove, *Rremove) *Rerror
	RemoveFile(*Tremovefile, *Rremove) *Rerror
	Stat(*Tstat, *Rstat) *Rerror
	Wstat(*Twstat, *Rwstat) *Rerror
	Renameat(*Trenameat, *Rrenameat) *Rerror
	GetFile(*Tgetfile, *Rread) *Rerror
	SetFile(*Tsetfile, *Rwrite) *Rerror
	PutFile(*Tputfile, *Rwrite) *Rerror
	WriteRead(*Twriteread, *Rread) *Rerror
	Detach(*Rdetach, DetachF) *Rerror
	Snapshot() []byte
}

type MkProtServer func(SessServer, fcall.Tsession) Protsrv
type RestoreProtServer func(SessServer, []byte) Protsrv

type DetachF func(fcall.Tsession)
