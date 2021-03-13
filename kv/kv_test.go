package kv

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"ulambda/fslib"
)

const NKEYS = 100

type Tstate struct {
	t    *testing.T
	s    *fslib.System
	fsl  *fslib.FsLib
	clrk *KvClerk
	ch   chan bool
}

func makeTstate(t *testing.T) *Tstate {
	ts := &Tstate{}
	ts.t = t
	ts.ch = make(chan bool)

	s, err := fslib.Boot("..")
	if err != nil {
		t.Fatalf("Boot %v\n", err)
	}
	ts.s = s
	ts.fsl = fslib.MakeFsLib("tester")

	err = ts.fsl.Mkdir("name/kv", 0777)
	if err != nil {
		t.Fatalf("Mkdir %v\n", err)
	}

	pid := ts.spawnKv()

	time.Sleep(1000 * time.Millisecond)

	pid1 := ts.spawnSharder("add", pid)
	ok, err := ts.fsl.Wait(pid1)
	assert.Nil(ts.t, err, "Wait")
	assert.Equal(t, string(ok), "OK")

	kc, err := MakeClerk()
	if err != nil {
		t.Fatalf("Make clerk %v\n", err)
	}
	ts.clrk = kc

	return ts
}

func (ts *Tstate) spawnKv() string {
	a := fslib.Attr{}
	a.Pid = fslib.GenPid()
	a.Program = "bin/kvd"
	a.Args = []string{}
	a.PairDep = nil
	a.ExitDep = nil
	ts.fsl.Spawn(&a)
	return a.Pid
}

func (ts *Tstate) spawnSharder(opcode, pid string) string {
	a := fslib.Attr{}
	a.Pid = fslib.GenPid()
	a.Program = "bin/sharderd"
	a.Args = []string{opcode, pid}
	a.PairDep = nil
	a.ExitDep = nil
	ts.fsl.Spawn(&a)
	return a.Pid
}

func (ts *Tstate) getKeys() {
	for i := 0; i < NKEYS; i++ {
		k := strconv.Itoa(i)
		v, err := ts.clrk.Get(k)
		assert.Nil(ts.t, err, "Get "+k)
		assert.Equal(ts.t, v, k, "Get")
	}
}

func (ts *Tstate) clerk() {
	for {
		select {
		case <-ts.ch:
			break
		default:
			ts.getKeys()
		}
	}
}

func TestConcur(t *testing.T) {
	ts := makeTstate(t)

	for i := 0; i < NKEYS; i++ {
		err := ts.clrk.Put(strconv.Itoa(i), strconv.Itoa(i))
		assert.Nil(t, err, "Put")
	}

	go ts.clerk()

	pids := make([]string, 0)
	for r := 0; r < NSHARD-1; r++ {
		pid := ts.spawnKv()
		pid1 := ts.spawnSharder("add", pid)
		ok, err := ts.fsl.Wait(pid1)
		assert.Nil(t, err, "Wait")
		assert.Equal(t, string(ok), "OK")
		time.Sleep(200 * time.Millisecond)
		pids = append(pids, pid)
	}

	for _, pid := range pids[1:] {
		pid1 := ts.spawnSharder("del", pid)
		ok, err := ts.fsl.Wait(pid1)
		assert.Nil(t, err, "Wait")
		assert.Equal(t, string(ok), "OK")
		time.Sleep(200 * time.Millisecond)
	}

	ts.ch <- true
	time.Sleep(200 * time.Millisecond)
	ts.s.Shutdown(ts.fsl)
}
