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
	rpcc  *rpcclnt.RPCClnt
	realm string
}

func NewDownloaddClnt(fsl *fslib.FsLib, srvId string, realm string) (*DownloaddClnt, error) {
	rpcc, err := rpcclnt.NewRPCClnt([]*fslib.FsLib{fsl}, path.Join(sp.DOWNLOADD, srvId))
	if err != nil {
		return nil, err
	}
	return &DownloaddClnt{
		FsLib: fsl,
		rpcc:  rpcc,
		realm: realm,
	}, nil
}

func (ddc *DownloaddClnt) DownloadLib(path string, copyFolder bool) error {
	req := &proto.DownloadLibRequest{
		NamedPath:  path,
		Realm:      ddc.realm,
		CopyFolder: copyFolder,
	}
	res := &proto.DownloadLibResponse{}
	if err := ddc.rpcc.RPC("Downloadd.DownloadLib", req, res); err != nil {
		return err
	}
	return nil
}

// func (ddc *DownloaddClnt) DownloadNamed(path string, copyFolder bool) error {
// 	req := &proto.DownloadNamedRequest{
// 		NamedPath:  path,
// 		Realm:      ddc.realm,
// 		CopyFolder: copyFolder,
// 	}
// 	res := &proto.DownloadNamedResponse{}
// 	if err := ddc.rpcc.RPC("Downloadd.DownloadNamed", req, res); err != nil {
// 		return err
// 	}
// 	return nil
// }
