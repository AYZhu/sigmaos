package downloaddclnt

import (
	"path"
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

func (ddc *DownloaddClnt) Download(path string) error {
	print("going to ask for download ")
	println(path)
	/**
	req := &proto.DownloadRequest{
		NamedPath: path,
	}
	res := &proto.DownloadResponse{}
	if err := ddc.rpcc.RPC("Downloadd.Download", req, res); err != nil {
		return err
	}*/
	return nil
}
