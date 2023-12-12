package example_test

import (
	// Go imports:
	"fmt"
	"log"
	"testing"
	"time"

	// External imports:
	"github.com/stretchr/testify/assert"

	// SigmaOS imports:
	"sigmaos/proc"
	sp "sigmaos/sigmap"
	"sigmaos/test"
)

func TestExerciseNamed(t *testing.T) {
	dir := sp.NAMED
	ts := test.NewTstatePath(t, dir)

	sts, err := ts.GetDir(dir)
	assert.Nil(t, err)

	log.Printf("%v: %v\n", dir, sp.Names(sts))

	// Your code here

	ts.Shutdown()
}

func TestExerciseS3(t *testing.T) {
	ts := test.NewTstateAll(t)

	// Your code here

	ts.Shutdown()
}

func TestExerciseProc(t *testing.T) {
	ts := test.NewTstateAll(t)

	p := proc.NewProc("example", []string{})
	start := time.Now()
	err := ts.Spawn(p)
	assert.Nil(t, err)
	duration := time.Since(start)
	err = ts.WaitStart(p.GetPid())
	assert.Nil(t, err)
	duration2 := time.Since(start)
	status, err := ts.WaitExit(p.GetPid())
	assert.Nil(t, err)
	duration3 := time.Since(start)
	assert.True(t, status.IsStatusOK())

	// Once you modified cmd/user/example, you should
	// pass this test:
	assert.Equal(t, "Hello world", status.Msg())
	fmt.Printf("cold spawn %v, start %v, exit %v", duration, duration2, duration3)

	p2 := proc.NewProc("example", []string{})
	start = time.Now()
	err = ts.Spawn(p2)
	assert.Nil(t, err)
	duration = time.Since(start)
	err = ts.WaitStart(p2.GetPid())
	assert.Nil(t, err)
	duration2 = time.Since(start)
	status, err = ts.WaitExit(p2.GetPid())
	assert.Nil(t, err)
	duration3 = time.Since(start)
	assert.True(t, status.IsStatusOK())

	// Once you modified cmd/user/example, you should
	// pass this test:
	assert.Equal(t, "Hello world", status.Msg())
	fmt.Printf("warm spawn %v, start %v, exit %v", duration, duration2, duration3)

	ts.Shutdown()
}

func TestExerciseParallel(t *testing.T) {
	ts := test.NewTstateAll(t)

	// Your code here

	ts.Shutdown()
}
