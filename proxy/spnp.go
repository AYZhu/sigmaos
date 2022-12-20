package proxy

import (
	"bufio"
	"bytes"
	"errors"
	"io"

	"sigmaos/fcall"
	"sigmaos/fs"
	np "sigmaos/ninep"
	"sigmaos/npcodec"
	"sigmaos/spcodec"
)

func Sp2NpDir(d []byte, cnt fcall.Tsize) ([]byte, *fcall.Err) {
	rdr := bytes.NewReader(d)
	brdr := bufio.NewReader(rdr)
	npsts := make([]*np.Stat9P, 0)
	for {
		spst, err := spcodec.UnmarshalDirEnt(brdr)
		if err != nil && errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		npst := npcodec.Sp2NpStat(spst)
		npsts = append(npsts, npst)
	}
	d, _, err := fs.MarshalDir(cnt, npsts)
	return d, err
}
