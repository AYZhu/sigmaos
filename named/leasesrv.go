package named

import (
	"context"

	"go.etcd.io/etcd/client/v3"

	db "sigmaos/debug"
	"sigmaos/fs"
	"sigmaos/fsetcd"
	leaseproto "sigmaos/lease/proto"
	sp "sigmaos/sigmap"
	"sigmaos/syncmap"
)

type LeaseSrv struct {
	lt *syncmap.SyncMap[sp.TclntId, clientv3.LeaseID]
	fs *fsetcd.FsEtcd
	lc clientv3.Lease
}

func newLeaseSrv(fs *fsetcd.FsEtcd) *LeaseSrv {
	return &LeaseSrv{
		lt: syncmap.NewSyncMap[sp.TclntId, clientv3.LeaseID](),
		fs: fs,
		lc: clientv3.NewLease(fs.Client),
	}
}

func (ls *LeaseSrv) AskLease(ctx fs.CtxI, req leaseproto.AskRequest, rep *leaseproto.AskResult) error {
	db.DPrintf(db.LEASESRV, "%v: AskLease %v\n", ctx.ClntId(), req.TTL)
	if lid, ok := ls.lt.Lookup(ctx.ClntId()); ok {
		rep.LeaseId = uint64(lid)
		return nil
	}
	resp, err := ls.lc.Grant(context.TODO(), int64(req.TTL))
	if err != nil {
		return err
	}
	rep.LeaseId = uint64(resp.ID)
	return nil
}

func (ls *LeaseSrv) Extend(ctx fs.CtxI, req leaseproto.ExtendRequest, rep *leaseproto.ExtendResult) error {
	db.DPrintf(db.LEASESRV, "%v: Extend %v\n", ctx.ClntId(), req.LeaseId)
	resp, err := ls.lc.KeepAliveOnce(context.TODO(), clientv3.LeaseID(req.LeaseId))
	if err != nil {
		return err
	}
	db.DPrintf(db.LEASESRV, "%v: Extend KeepAliveOnce [%v,%v]\n", ctx.ClntId(), resp.ID, resp.TTL)
	return nil
}

func (ls *LeaseSrv) End(ctx fs.CtxI, req leaseproto.ExtendRequest, rep *leaseproto.ExtendResult) error {
	db.DPrintf(db.LEASESRV, "%v: End %v\n", ctx.ClntId(), req.LeaseId)
	resp, err := ls.lc.Revoke(context.TODO(), clientv3.LeaseID(req.LeaseId))
	if err != nil {
		return err
	}
	db.DPrintf(db.LEASESRV, "%v: End Revoke %v\n", ctx.ClntId(), resp)
	return nil
}
