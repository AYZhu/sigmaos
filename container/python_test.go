package container_test

import (
	"fmt"
	"sigmaos/proc"
	"sigmaos/test"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPythonLaunch(t *testing.T) {
	ts := test.NewTstateAll(t)
	p := proc.NewProc("python", []string{"/~~/pyproc/hello.py"})
	start := time.Now()
	err := ts.Spawn(p)
	assert.Nil(ts.T, err)
	duration := time.Since(start)
	err = ts.WaitStart(p.GetPid())
	assert.Nil(ts.T, err, "Error waitstart: %v", err)
	duration2 := time.Since(start)
	st, err := ts.WaitExit(p.GetPid())
	assert.Nil(t, err)
	duration3 := time.Since(start)
	fmt.Printf("cold spawn %v, start %v, exit %v", duration, duration2, duration3)

	p2 := proc.NewProc("python", []string{"/~~/pyproc/hello.py"})
	start = time.Now()
	err = ts.Spawn(p2)
	assert.Nil(ts.T, err)
	duration = time.Since(start)
	err = ts.WaitStart(p2.GetPid())
	assert.Nil(ts.T, err, "Error waitstart: %v", err)
	duration2 = time.Since(start)
	st, err = ts.WaitExit(p2.GetPid())
	assert.Nil(t, err)
	duration3 = time.Since(start)
	assert.True(t, st.IsStatusOK(), st)
	fmt.Printf("warm spawn %v, start %v, exit %v", duration, duration2, duration3)

	ts.Shutdown()
}
func TestPythonLib(t *testing.T) {
	ts := test.NewTstateAll(t)
	p := proc.NewProc("python", []string{"/~~/pyproc/stdimports.py"})
	err := ts.Spawn(p)
	err = ts.WaitStart(p.GetPid())
	st, err := ts.WaitExit(p.GetPid())
	assert.Nil(t, err)
	assert.True(t, st.IsStatusOK(), st)
	ts.Shutdown()
}
