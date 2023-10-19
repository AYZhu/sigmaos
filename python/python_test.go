package python_test

import (
	// Go imports:

	"testing"

	// External imports:
	"github.com/stretchr/testify/assert"

	// SigmaOS imports:
	"sigmaos/proc"
	sp "sigmaos/sigmap"
	"sigmaos/spproxy"
	"sigmaos/test"
)

func TestExerciseProc(t *testing.T) {
	ts := test.MakeTstateAll(t)

	spproxy.RunSPProxySrv(true)

	fd, err := ts.Create("name/tfile.py", sp.DMWRITE|sp.DMREAD, sp.OWRITE)
	assert.Nil(t, err)
	ts.Write(fd, []byte("print(\"hello, world!\")\n"))
	ts.Close(fd)

	p := proc.MakeProc("python", []string{"name/tfile.py"})
	err = ts.Spawn(p)
	assert.Nil(t, err)
	err = ts.WaitStart(p.GetPid())
	assert.Nil(t, err)
	status, err := ts.WaitExit(p.GetPid())
	assert.Nil(t, err)
	assert.True(t, status.IsStatusOK())

	err = ts.Remove("name/tfile.py")
	assert.Nil(t, err)

	// Once you modified cmd/user/example, you should
	// pass this test:
	assert.Equal(t, "Hello world", status.Msg())

	ts.Shutdown()
}
