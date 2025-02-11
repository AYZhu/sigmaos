package named_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	db "sigmaos/debug"
	"sigmaos/fslib"
	"sigmaos/named"
	sp "sigmaos/sigmap"
	"sigmaos/test"
)

func TestBootNamed(t *testing.T) {
	ts := test.MakeTstateAll(t)

	sts, err := ts.GetDir(sp.NAMED + "/")
	assert.Nil(t, err)
	db.DPrintf(db.TEST, "named %v\n", sp.Names(sts))

	assert.True(t, fslib.Present(sts, named.InitRootDir), "initfs")

	ts.Shutdown()
}

func TestKillNamed(t *testing.T) {
	ts := test.MakeTstateAll(t)

	sts, err := ts.GetDir(sp.NAMED + "/")
	assert.Nil(t, err)

	db.DPrintf(db.TEST, "named %v\n", sp.Names(sts))

	err = ts.Boot(sp.NAMEDREL)
	assert.Nil(t, err)

	sts, err = ts.GetDir(sp.NAMED + "/")
	assert.Nil(t, err)
	db.DPrintf(db.TEST, "named %v\n", sp.Names(sts))

	db.DPrintf(db.TEST, "kill named..\n")

	err = ts.KillOne(sp.NAMEDREL)
	assert.Nil(t, err)

	db.DPrintf(db.TEST, "GetDir..\n")

	sts, err = ts.GetDir(sp.NAMED + "/")
	assert.Nil(t, err)
	db.DPrintf(db.TEST, "named %v\n", sp.Names(sts))

	ts.Shutdown()
}
