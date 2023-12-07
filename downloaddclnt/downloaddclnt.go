package downloaddclnt

import (
	"path"
	"sigmaos/downloadd/proto"
	"sigmaos/fslib"
	"sigmaos/rpcclnt"
	sp "sigmaos/sigmap"
)

type DownloaddClnt struct {
	*fslib.FsLib
	rpcc *rpcclnt.RPCClnt
}

func NewDownloaddClnt(fsl *fslib.FsLib, srvId string) (*DownloaddClnt, error) {
	rpcc, err := rpcclnt.NewRPCClnt([]*fslib.FsLib{fsl}, path.Join(sp.DOWNLOADD, srvId))
	if err != nil {
		return nil, err
	}
	return &DownloaddClnt{
		FsLib: fsl,
		rpcc:  rpcc,
	}, nil
}

func (ddc *DownloaddClnt) Download(path string, realm string) error {
	req := &proto.DownloadLibRequest{
		NamedPath: path,
		Realm:     realm,
	}
	res := &proto.DownloadLibResponse{}
	if err := ddc.rpcc.RPC("Downloadd.Download", req, res); err != nil {
		return err
	}
	print("got back ")
	println(res.GetTmpPath())
	return nil
}
