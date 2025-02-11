package leaderclnt

import (
	db "sigmaos/debug"
	"sigmaos/electclnt"
	"sigmaos/fenceclnt"
	"sigmaos/fslib"
	"sigmaos/proc"
	sp "sigmaos/sigmap"
)

//
// Library for becoming a leader for an epoch.
//

type LeaderClnt struct {
	fc *fenceclnt.FenceClnt
	ec *electclnt.ElectClnt
	pn string
}

func MakeLeaderClnt(fsl *fslib.FsLib, pn string, perm sp.Tperm) (*LeaderClnt, error) {
	l := &LeaderClnt{pn: pn, fc: fenceclnt.MakeFenceClnt(fsl)}
	ec, err := electclnt.MakeElectClnt(fsl, pn, perm)
	if err != nil {
		return nil, err
	}
	l.ec = ec
	return l, nil
}

// Become leader and fence ops at that epoch.  Another proc may steal
// our leadership (e.g., after we are partioned) and start a higher
// epoch.  Note epoch doesn't take effect until we perform a fenced
// operation (e.g., a read/write).
func (l *LeaderClnt) LeadAndFence(b []byte, dirs []string) error {
	if err := l.ec.AcquireLeadership(b); err != nil {
		return err
	}
	db.DPrintf(db.LEADER, "%v: LeadAndFence: %v\n", proc.GetName(), l.Fence())
	return l.fenceDirs(dirs)
}

func (l *LeaderClnt) Fence() sp.Tfence {
	return l.ec.Fence()
}

func (l *LeaderClnt) fenceDirs(dirs []string) error {
	if err := l.fc.FenceAtEpoch(l.Fence(), dirs); err != nil {
		return err
	}
	return nil
}

func (l *LeaderClnt) GetFences(pn string) ([]*sp.Stat, error) {
	return l.fc.GetFences(pn)
}

// Works for file systems that support fencefs
func (l *LeaderClnt) RemoveFence(dirs []string) error {
	return l.fc.RemoveFence(dirs)
}

func (l *LeaderClnt) ReleaseLeadership() error {
	return l.ec.ReleaseLeadership()
}

func (l *LeaderClnt) Lease() sp.TleaseId {
	return l.ec.Lease()
}
