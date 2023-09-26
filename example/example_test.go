package example_test

import (
	// Go imports:
	"bufio"
	"log"
	"testing"

	// External imports:
	"github.com/stretchr/testify/assert"

	// SigmaOS imports:
	"sigmaos/proc"
	sp "sigmaos/sigmap"
	"sigmaos/test"
)

func TestExerciseNamed(t *testing.T) {
	dir := sp.NAMED
	ts := test.MakeTstatePath(t, dir)

	sts, err := ts.GetDir(dir)
	assert.Nil(t, err)

	log.Printf("%v: %v\n", dir, sp.Names(sts))

	// Your code here

	fd, err := ts.Create("name/tfile", sp.DMWRITE|sp.DMREAD, sp.OWRITE)
	assert.Nil(t, err)
	ts.Write(fd, []byte("hello, world!\n"))
	ts.Close(fd)

	sts, err = ts.GetDir(dir)
	assert.Nil(t, err)

	log.Printf("%v: %v\n", dir, sp.Names(sts))
	assert.Contains(t, sp.Names(sts), "tfile")

	var length (sp.Tsize)

	for _, st := range sts {
		if st.Name == "tfile" {
			length = sp.Tsize(st.Length)
		}
	}

	fd, err = ts.Open("name/tfile", sp.OREAD)
	assert.Nil(t, err)
	out, err := ts.Read(fd, length)
	assert.Nil(t, err)

	err = ts.Remove("name/tfile")
	assert.Nil(t, err)

	assert.Equal(t, []byte("hello, world!\n"), out)
	ts.Close(fd)

	ts.Shutdown()
}

func TestExerciseS3(t *testing.T) {
	ts := test.MakeTstateAll(t)

	rdr, err := ts.OpenReader("name/s3/~any/sigmaos-alanyzhu/gutenberg/pg-tom_sawyer.txt")
	assert.Nil(t, err)

	scanner := bufio.NewScanner(rdr)
	scanner.Split(bufio.ScanWords)

	count := 0

	for scanner.Scan() {
		if scanner.Text() == "the" {
			count += 1
		}
	}

	assert.Equal(t, 3481, count)

	ts.Shutdown()
}

func TestExerciseProc(t *testing.T) {
	ts := test.MakeTstateAll(t)

	p := proc.MakeProc("example", []string{})
	err := ts.Spawn(p)
	assert.Nil(t, err)
	err = ts.WaitStart(p.GetPid())
	assert.Nil(t, err)
	status, err := ts.WaitExit(p.GetPid())
	assert.Nil(t, err)
	assert.True(t, status.IsStatusOK())

	// Once you modified cmd/user/example, you should
	// pass this test:
	assert.Equal(t, "Hello world", status.Msg())

	ts.Shutdown()
}

func MakeProc(t *testing.T, ts *test.Tstate, path string, ch chan int) {
	p := proc.MakeProc("example", []string{path})
	err := ts.Spawn(p)
	assert.Nil(t, err)
	err = ts.WaitStart(p.GetPid())
	assert.Nil(t, err)
	status, err := ts.WaitExit(p.GetPid())
	assert.Nil(t, err)
	assert.True(t, status.IsStatusOK())
	ch <- int(status.StatusData.(float64))
}

func TestExerciseParallel(t *testing.T) {
	ts := test.MakeTstateAll(t)

	sts, err := ts.GetDir("name/s3/~any/sigmaos-alanyzhu/gutenberg/")
	assert.Nil(t, err)

	chs := []chan int{}

	for _, st := range sts {
		ch := make(chan int)
		chs = append(chs, ch)
		go MakeProc(t, ts, "name/s3/~any/sigmaos-alanyzhu/gutenberg/"+st.Name, ch)
	}

	outs := 0
	for _, ch := range chs {
		outs = outs + <-ch
	}

	assert.Equal(t, 59220, outs)
	ts.Shutdown()
}
