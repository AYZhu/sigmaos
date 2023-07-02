package leaderclnt

import (
	"sigmaos/electclnt"
	"sigmaos/epochclnt"
	"sigmaos/fenceclnt"
	"sigmaos/fslib"
	"sigmaos/sessp"
	sp "sigmaos/sigmap"
)

//
// Library for becoming a leader for an epoch.
//

type LeaderClnt struct {
	*fslib.FsLib
	*epochclnt.EpochClnt
	*fenceclnt.FenceClnt
	epochfn string
	e       *electclnt.ElectClnt
}

func MakeLeaderClnt(fsl *fslib.FsLib, leaderfn string, perm sp.Tperm) *LeaderClnt {
	l := &LeaderClnt{}
	l.FsLib = fsl
	l.e = electclnt.MakeElectClnt(fsl, leaderfn, perm)
	l.EpochClnt = epochclnt.MakeEpochClnt(fsl, leaderfn, perm)
	l.FenceClnt = fenceclnt.MakeFenceClnt(fsl, l.EpochClnt)
	return l
}

func (l *LeaderClnt) EpochPath() string {
	return l.epochfn
}

// Become leader for an epoch and fence ops for that epoch.  Another
// proc may steal our leadership (e.g., after we are partioned) and
// start a higher epoch.  Note epoch doesn't take effect until we
// perform a fenced operation (e.g., a read/write).
func (l *LeaderClnt) AcquireFencedEpoch(b []byte, dirs []string) (sessp.Tepoch, error) {
	if err := l.e.AcquireLeadership(b); err != nil {
		return sessp.NoEpoch, err
	}
	return l.EnterNextEpoch(dirs)
}

// Enter next epoch.  If the leader is partitioned and another leader
// has taken over, this fails.
func (l *LeaderClnt) EnterNextEpoch(dirs []string) (sessp.Tepoch, error) {
	epoch, err := l.AdvanceEpoch()
	if err != nil {
		return sessp.NoEpoch, err
	}
	if err := l.FenceAtEpoch(epoch, dirs); err != nil {
		return sessp.NoEpoch, err
	}
	return epoch, nil
}
