package leaseclnt

import (
	db "sigmaos/debug"
	"sigmaos/fsetcd"
	"sigmaos/fslib"
	leaseproto "sigmaos/lease/proto"
	"sigmaos/proc"
	"sigmaos/rpcclnt"
	sp "sigmaos/sigmap"
	"sigmaos/syncmap"
)

type LeaseClnt struct {
	*fslib.FsLib
	lm *syncmap.SyncMap[string, *LeaseInfo]
	cc *rpcclnt.ClntCache
}

func NewLeaseClnt(fsl *fslib.FsLib) (*LeaseClnt, error) {
	return &LeaseClnt{
		FsLib: fsl,
		lm:    syncmap.NewSyncMap[string, *LeaseInfo](),
		cc:    rpcclnt.NewRPCClntCache([]*fslib.FsLib{fsl}),
	}, nil
}

// Ask for lease; if caller already has a lease at that server, return
// it.
func (lmc *LeaseClnt) AskLease(pn string, ttl sp.Tttl) (*LeaseInfo, error) {
	srv, rest, err := lmc.PathLastSymlink(pn)
	db.DPrintf(db.LEASECLNT, "AskLease %v: %v %v err %v\n", pn, srv, rest, err)
	if li, ok := lmc.lm.Lookup(srv.String()); ok {
		return li, nil
	}
	var res leaseproto.AskResult
	if err := lmc.cc.RPC(srv.String(), "LeaseSrv.AskLease", &leaseproto.AskRequest{
		ClntId: uint64(lmc.ClntId()),
		TTL:    fsetcd.LeaseTTL}, &res); err == nil {
		li := &LeaseInfo{
			ch:  make(chan struct{}),
			srv: srv.String(),
			lid: sp.TleaseId(res.LeaseId),
			lmc: lmc,
		}
		db.DPrintf(db.LEASECLNT, "AskLease %q %v\n", srv, li)
		lmc.lm.Insert(srv.String(), li)
		return li, nil
	} else {
		db.DPrintf(db.LEASECLNT, "AskLease %v err %v\n", pn, err)
		return nil, err
	}
}

func (lmgr *LeaseClnt) EndLeases() error {
	db.DPrintf(db.LEASECLNT, "%v: EndLeases\n", proc.GetPid())
	for _, li := range lmgr.lm.Values() {
		db.DPrintf(db.LEASECLNT, "%v: EndLeases %v\n", proc.GetPid(), li)
		li.End()
	}
	return nil
}
