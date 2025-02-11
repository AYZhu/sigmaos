package realmclnt

import (
	"sigmaos/fslib"
	"sigmaos/realmsrv/proto"
	"sigmaos/rpcclnt"
	sp "sigmaos/sigmap"
)

type RealmClnt struct {
	*fslib.FsLib
	rpcc *rpcclnt.RPCClnt
}

func MakeRealmClnt(fsl *fslib.FsLib) (*RealmClnt, error) {
	rc := &RealmClnt{FsLib: fsl}
	rpcc, err := rpcclnt.MkRPCClnt([]*fslib.FsLib{rc.FsLib}, sp.REALMD)
	if err != nil {
		return nil, err
	}
	rc.rpcc = rpcc
	return rc, nil
}

func (rc *RealmClnt) MakeRealm(realm sp.Trealm, net string) error {
	req := &proto.MakeRequest{
		Realm:   realm.String(),
		Network: net,
	}
	res := &proto.MakeResult{}
	if err := rc.rpcc.RPC("RealmSrv.Make", req, res); err != nil {
		return err
	}
	return nil
}

func (rc *RealmClnt) RemoveRealm(realm sp.Trealm) error {
	req := &proto.RemoveRequest{
		Realm: realm.String(),
	}
	res := &proto.RemoveResult{}
	if err := rc.rpcc.RPC("RealmSrv.Remove", req, res); err != nil {
		return err
	}
	return nil
}
