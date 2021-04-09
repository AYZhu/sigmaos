package kv

import (
	"log"
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
	pid  string
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
	ts.fsl = fslib.MakeFsLib("kv_test")

	err = ts.fsl.Mkdir("name/kv", 07)
	if err != nil {
		t.Fatalf("Mkdir %v\n", err)
	}

	ts.pid = ts.spawnKv()

	time.Sleep(1000 * time.Millisecond)

	pid1 := ts.spawnSharder("add", kvname(ts.pid))
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

func key(k int) string {
	return "key" + strconv.Itoa(k)

}

func (ts *Tstate) getKeys() bool {
	for i := 0; i < NKEYS; i++ {
		v, err := ts.clrk.Get(key(i))
		select {
		case <-ts.ch:
			return true
		default:
			assert.Nil(ts.t, err, "Get "+key(i))
			assert.Equal(ts.t, key(i), v, "Get")
		}
	}
	return false
}

func (ts *Tstate) clerk() {
	done := false
	for !done {
		done = ts.getKeys()
	}
}

func TestConcur(t *testing.T) {
	ts := makeTstate(t)

	for i := 0; i < NKEYS; i++ {
		err := ts.clrk.Put(key(i), key(i))
		assert.Nil(t, err, "Put")
	}

	go ts.clerk()

	pids := make([]string, 0)
	// for r := 0; r < 1; r++ {
	for r := 0; r < NSHARD-1; r++ {
		pid := ts.spawnKv()
		log.Printf("Add %v\n", pid)
		pid1 := ts.spawnSharder("add", kvname(pid))
		log.Printf("wait")
		ok, err := ts.fsl.Wait(pid1)
		assert.Nil(t, err, "Wait")
		assert.Equal(t, string(ok), "OK")
		log.Printf("wait done")
		time.Sleep(200 * time.Millisecond)
		pids = append(pids, pid)
	}

	for _, pid := range pids {
		log.Printf("Del %v\n", pid)
		pid1 := ts.spawnSharder("del", kvname(pid))
		ok, err := ts.fsl.Wait(pid1)
		assert.Nil(t, err, "Wait")
		assert.Equal(t, string(ok), "OK")
		time.Sleep(200 * time.Millisecond)
	}

	// stop clerk
	ts.ch <- true

	// delete first KV
	pid1 := ts.spawnSharder("del", ts.pid)
	ok, err := ts.fsl.Wait(pid1)
	assert.Nil(t, err, "Wait")
	assert.Equal(t, string(ok), "OK")
	time.Sleep(200 * time.Millisecond)

	ts.s.Shutdown(ts.fsl)
}
