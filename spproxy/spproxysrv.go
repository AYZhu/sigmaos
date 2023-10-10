package spproxy

import (
	dbg "sigmaos/debug"
	"sigmaos/fs"
	"sigmaos/rand"
	"sigmaos/sigmaclnt"
	sp "sigmaos/sigmap"
	"sigmaos/sigmasrv"
)

// YH:
// Toy server echoing request message

type SigmapSrv struct {
	sid string
}

const DEBUG_SP_SERVER = "SPPROXYSRV"
const DIR_SP_SERVER = sp.NAMED + "spproxysrv/"
const NAMED_SP_SERVER = DIR_SP_SERVER + "spproxy-server"

func RunSPProxySrv(public bool) error {
	sigmapsrv := &SigmapSrv{rand.String(8)}
	dbg.DPrintf(DEBUG_SP_SERVER, "==%v== Creating spproxysrv server \n", sigmapsrv.sid)
	ssrv, err := sigmasrv.MakeSigmaSrvPublic(NAMED_SP_SERVER, sigmapsrv, DEBUG_SP_SERVER, public)
	if err != nil {
		return err
	}
	dbg.DPrintf(DEBUG_SP_SERVER, "==%v== Starting to run spproxy service\n", sigmapsrv.sid)
	return ssrv.RunServer()
}

// find meaning of life for request
func (sigmapsrv *SigmapSrv) SPOpen(ctx fs.CtxI, req *OpenRequest, rep *OpenResult) error {
	dbg.DPrintf(DEBUG_SP_SERVER, "==%v== Received Echo Request: %v\n", sigmapsrv.sid, req)
	fs, err := sigmaclnt.MkSigmaClntFsLib(ctx.Uname())
	if err != nil {
		return err
	}
	fd, err := fs.Open(req.Text, sp.OWRITE) // TODO: permissioning!

	if err != nil {
		return err
	}

	rep.Result = int64(fd)

	return nil
}
