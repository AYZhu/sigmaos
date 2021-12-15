package fslib_test

import (
	"log"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"ulambda/fslib"
	"ulambda/kernel"
	"ulambda/named"
	np "ulambda/ninep"
)

type Tstate struct {
	*fslib.FsLib
	t *testing.T
	s *kernel.System
}

func makeTstate(t *testing.T) *Tstate {
	ts := &Tstate{}
	ts.s = kernel.MakeSystemNamed("..")
	ts.FsLib = fslib.MakeFsLibAddr("fslibtest", fslib.Named())
	ts.t = t
	return ts
}

func TestRemoveSimple(t *testing.T) {
	ts := makeTstate(t)

	fn := "name/f"
	d := []byte("hello")
	err := ts.MakeFile(fn, 0777, np.OWRITE, d)
	assert.Equal(t, nil, err)

	err = ts.Remove(fn)
	assert.Equal(t, nil, err)

	_, err = ts.Stat(fn)
	assert.NotEqual(t, nil, err)

	ts.s.Shutdown()
}

func TestRemoveNonExistent(t *testing.T) {
	ts := makeTstate(t)

	fn := "name/f"
	d := []byte("hello")
	err := ts.MakeFile(fn, 0777, np.OWRITE, d)
	assert.Equal(t, nil, err)

	err = ts.Remove("name/this-file-does-not-exist")
	assert.NotNil(t, err)

	ts.s.Shutdown()
}

func TestRemovePath(t *testing.T) {
	ts := makeTstate(t)

	err := ts.Mkdir("name/d1", 0777)
	assert.Equal(t, nil, err)
	fn := "name/d1/f"
	d := []byte("hello")
	err = ts.MakeFile(fn, 0777, np.OWRITE, d)
	assert.Equal(t, nil, err)

	d1, err := ts.ReadFile(fn)
	assert.Equal(t, "hello", string(d1))

	err = ts.Remove(fn)
	assert.Equal(t, nil, err)

	ts.s.Shutdown()
}

func TestRename(t *testing.T) {
	ts := makeTstate(t)
	err := ts.Mkdir("name/d1", 0777)
	assert.Equal(t, nil, err)
	err = ts.Mkdir("name/d2", 0777)
	assert.Equal(t, nil, err)

	fn := "name/d1/f"
	fn1 := "name/d2/g"
	d := []byte("hello")
	err = ts.MakeFile(fn, 0777, np.OWRITE, d)
	assert.Equal(t, nil, err)

	err = ts.Rename(fn, fn1)
	assert.Equal(t, nil, err)

	d1, err := ts.ReadFile(fn1)
	assert.Equal(t, "hello", string(d1))
	ts.s.Shutdown()
}

func TestRenameAndRemove(t *testing.T) {
	ts := makeTstate(t)
	err := ts.Mkdir("name/d1", 0777)
	assert.Equal(t, nil, err)
	err = ts.Mkdir("name/d2", 0777)
	assert.Equal(t, nil, err)

	fn := "name/d1/f"
	fn1 := "name/d2/g"
	d := []byte("hello")
	err = ts.MakeFile(fn, 0777, np.OWRITE, d)
	assert.Equal(t, nil, err)

	err = ts.Rename(fn, fn1)
	assert.Equal(t, nil, err)

	d1, err := ts.ReadFile(fn1)
	assert.Equal(t, nil, err)
	assert.Equal(t, "hello", string(d1))

	_, err = ts.Stat(fn1)
	assert.Equal(t, nil, err)

	err = ts.Remove(fn1)
	assert.Equal(t, nil, err)
	ts.s.Shutdown()
}

func TestNonEmpty(t *testing.T) {
	ts := makeTstate(t)
	err := ts.Mkdir("name/d1", 0777)
	assert.Equal(t, nil, err)
	err = ts.Mkdir("name/d2", 0777)
	assert.Equal(t, nil, err)

	fn := "name/d1/f"
	d := []byte("hello")
	err = ts.MakeFile(fn, 0777, np.OWRITE, d)
	assert.Equal(t, nil, err)

	err = ts.Remove("name/d1")
	assert.NotNil(t, err, "Remove")

	err = ts.Rename("name/d2", "name/d1")
	assert.NotNil(t, err, "Rename")

	ts.s.Shutdown()
}

func TestCopy(t *testing.T) {
	ts := makeTstate(t)
	d := []byte("hello")
	src := "name/f"
	dst := "name/g"
	err := ts.MakeFile(src, 0777, np.OWRITE, d)
	assert.Equal(t, nil, err)

	err = ts.CopyFile(src, dst)
	assert.Equal(t, nil, err)

	d1, err := ts.ReadFile(dst)
	assert.Equal(t, "hello", string(d1))

	ts.s.Shutdown()
}

func TestDirSimple(t *testing.T) {
	ts := makeTstate(t)
	dn := "name/d"
	err := ts.Mkdir(dn, 0777)
	assert.Equal(t, nil, err)
	b, err := ts.IsDir(dn)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, b)

	d := []byte("hello")
	err = ts.MakeFile(dn+"/f", 0777, np.OWRITE, d)
	assert.Equal(t, nil, err)

	sts, err := ts.ReadDir(dn)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(sts))
	assert.Equal(t, "f", sts[0].Name)

	err = ts.RmDir(dn)
	_, err = ts.Stat(dn)
	assert.NotEqual(t, nil, err)

	ts.s.Shutdown()
}

func TestDirDot(t *testing.T) {
	ts := makeTstate(t)
	dn := "name/dir0"
	err := ts.Mkdir(dn, 0777)
	assert.Equal(t, nil, err)
	b, err := ts.IsDir(dn + "/.")
	assert.Equal(t, nil, err)
	assert.Equal(t, true, b)
	err = ts.RmDir(dn + "/.")
	assert.NotEqual(t, nil, err)
	err = ts.RmDir(dn)
	_, err = ts.Stat(dn + "/.")
	assert.NotEqual(t, nil, err)
	_, err = ts.Stat("name/.")
	assert.Equal(t, nil, err)
	ts.s.Shutdown()
}

func (ts *Tstate) procdName(t *testing.T, exclude map[string]bool) string {
	sts, err := ts.ReadDir(named.PROCD)
	stsExcluded := []*np.Stat{}
	for _, s := range sts {
		if ok := exclude[path.Join(named.PROCD, s.Name)]; !ok {
			stsExcluded = append(stsExcluded, s)
		}
	}
	assert.Nil(t, err, named.PROCD)
	assert.Equal(t, 1, len(stsExcluded))
	name := path.Join(named.PROCD, stsExcluded[0].Name)
	return name
}

func TestCounter(t *testing.T) {
	const N = 10

	ts := makeTstate(t)
	fd, err := ts.CreateFile("name/cnt", 0777|np.DMTMP, np.OWRITE)
	assert.Equal(t, nil, err)
	b := []byte(strconv.Itoa(0))
	_, err = ts.Write(fd, b)
	assert.Equal(t, nil, err)
	err = ts.Close(fd)
	assert.Equal(t, nil, err)

	ch := make(chan int)

	for i := 0; i < N; i++ {
		go func(i int) {
			ntrial := 0
			for {
				ntrial += 1
				b, v, err := ts.GetFile("name/cnt")
				assert.Equal(t, nil, err)
				n, err := strconv.Atoi(string(b))
				assert.Equal(t, nil, err)
				n += 1
				b = []byte(strconv.Itoa(n))
				_, err = ts.SetFile("name/cnt", b, v)
				if err != nil && err.Error() == "Version mismatch" {
					continue
				}
				assert.Equal(t, nil, err)
				break
			}
			// log.Printf("%d: tries %v\n", i, ntrial)
			ch <- i
		}(i)
	}
	for i := 0; i < N; i++ {
		<-ch
	}
	fd, err = ts.Open("name/cnt", np.ORDWR)
	assert.Equal(t, nil, err)
	b, err = ts.Read(fd, 100)
	assert.Equal(t, nil, err)
	n, err := strconv.Atoi(string(b))
	assert.Equal(t, nil, err)

	assert.Equal(t, N, n)

	ts.s.Shutdown()
}

// Inline Set() so that we can delay the Write() to emulate a delay on
// the server between open and write.
func writeFile(fl *fslib.FsLib, fn string, d []byte) error {
	fd, err := fl.Open(fn, np.OWRITE)
	if err != nil {
		return err
	}
	time.Sleep(1000 * time.Nanosecond)
	_, err = fl.Write(fd, d)
	if err != nil {
		return err
	}
	err = fl.Close(fd)
	if err != nil {
		return err
	}
	return nil
}

// Test race: write returns successfully after rename, but read sees
// an old value,
func TestSetRenameGet(t *testing.T) {
	const N = 20_000

	ts := makeTstate(t)

	err := ts.Mkdir("name/d1", 0777)
	assert.Equal(t, nil, err)
	fn := "name/d1/f"
	fn1 := "name/d1/f1"
	d := []byte(strconv.Itoa(0))
	err = ts.MakeFile(fn, 0777, np.OWRITE, d)
	assert.Equal(t, nil, err)

	start := make(chan bool)
	ch := make(chan int)
	go func() {
		fsl := fslib.MakeFsLibAddr("fsl1", fslib.Named())
		for i := 1; i < N; {
			d := []byte(strconv.Itoa(i))
			err = writeFile(fsl, fn, d)
			if err == nil {
				i++
			} else {
				ch <- i - 1
				<-start
			}
		}
		ch <- N - 1
	}()

	race := false
	for true {
		err = ts.Rename(fn, fn1)
		assert.Equal(t, nil, err)

		d1, err := ts.ReadFile(fn1)
		n, err := strconv.Atoi(string(d1))
		assert.Equal(t, nil, err)

		m := <-ch

		if n != m {
			log.Printf("%v %v\n", m, n)
			race = true
		}
		if m == N-1 {
			break
		}

		err = ts.Rename(fn1, fn)
		assert.Equal(t, nil, err)

		start <- true
	}
	assert.Equal(ts.t, true, race, "SetRenameGet")

	ts.s.Shutdown()
}

func TestWatchCreate(t *testing.T) {
	ts := makeTstate(t)

	fn := "name/w"
	ch := make(chan bool)
	_, err := ts.ReadFileWatch(fn, func(string, error) {
		ch <- true
	})
	assert.NotEqual(t, nil, err)
	if err != nil {
		assert.Equal(t, true, strings.HasPrefix(err.Error(), "file not found"))
	}

	err = ts.MakeFile(fn, 0777, np.OWRITE, nil)
	assert.Equal(t, nil, err)

	<-ch

	ts.s.Shutdown()
}

func TestWatchRemoveSeq(t *testing.T) {
	ts := makeTstate(t)

	fn := "name/w"
	err := ts.MakeFile(fn, 0777, np.OWRITE, nil)
	assert.Equal(t, nil, err)

	ch := make(chan bool)
	err = ts.SetRemoveWatch(fn, func(string, error) {
		ch <- true
	})

	err = ts.Remove(fn)
	assert.Equal(t, nil, err)

	<-ch

	ts.s.Shutdown()
}

func TestWatchDir(t *testing.T) {
	ts := makeTstate(t)

	fn := "name/d1"
	err := ts.Mkdir(fn, 0777)
	assert.Equal(t, nil, err)

	ch := make(chan bool)
	err = ts.SetDirWatch(fn, func(string, error) {
		ch <- true
	})
	assert.Equal(t, nil, err)

	err = ts.MakeFile(fn+"/x", 0777, np.OWRITE, nil)
	assert.Equal(t, nil, err)

	<-ch

	ts.s.Shutdown()
}

func TestLock1(t *testing.T) {
	ts := makeTstate(t)
	ch := make(chan int)
	ts.Mkdir("name/locks", 0777)

	// Lock the file
	err := ts.MakeFile("name/locks/test-lock", 0777|np.DMTMP, np.OWRITE|np.OCEXEC, []byte{})
	assert.Equal(t, nil, err)
	fsl := fslib.MakeFsLibAddr("fslibtest0", fslib.Named())
	go func() {
		err := fsl.MakeFile("name/locks/test-lock", 0777|np.DMTMP, np.OWRITE|np.OWATCH, []byte{})
		assert.Nil(t, err, "MakeFile")
		ch <- 0
	}()
	time.Sleep(time.Second * 2)
	err = ts.Remove("name/locks/test-lock")
	assert.Nil(t, err, "Remove")
	go func() {
		time.Sleep(2 * time.Second)
		ch <- 1
	}()
	i := <-ch
	assert.Equal(t, 0, i)

	ts.s.Shutdown()
}

func TestLockN(t *testing.T) {
	const N = 20

	ts := makeTstate(t)
	ch := make(chan int)
	acquired := false
	for i := 0; i < N; i++ {
		go func(i int) {
			fsl := fslib.MakeFsLibAddr("fslibtest"+strconv.Itoa(i), fslib.Named())
			err := fsl.MakeFile("name/lock", 0777|np.DMTMP, np.OWRITE|np.OWATCH, []byte{})
			assert.Equal(t, nil, err)
			assert.Equal(t, false, acquired)
			acquired = true
			ch <- i
		}(i)
	}
	for i := 0; i < N; i++ {
		<-ch
		// log.Printf("%d acquired lock\n", j)
		acquired = false
		err := ts.Remove("name/lock")
		assert.Equal(t, nil, err)
	}
	ts.s.Shutdown()
}

func TestLockAfterConnClose(t *testing.T) {
	ts := makeTstate(t)

	lPath := "name/lock-conn-close-test"

	fsl1 := fslib.MakeFsLibAddr("fslibtest-1", fslib.Named())

	err := ts.MakeFile(lPath, 0777|np.DMTMP, np.OWRITE|np.OWATCH, []byte{})
	assert.Nil(t, err, "Make lock 1")

	go func() {
		// Should wait
		err := fsl1.MakeFile(lPath, 0777|np.DMTMP, np.OWRITE|np.OWATCH, []byte{})
		assert.Equal(t, err.Error(), "EOF", "Make lock 2")
	}()

	time.Sleep(500 * time.Millisecond)

	// Kill fsl2's connection
	fsl1.Exit()

	// Remove the lock file
	ts.Remove(lPath)
	assert.Equal(t, nil, err)

	// Try to lock again (should succeed)
	err = ts.MakeFile(lPath, 0777|np.DMTMP, np.OWRITE|np.OWATCH, []byte{})
	assert.Nil(t, err, "Make lock 3")

	ts.s.Shutdown()
}

// Test race: write returns successfully after rename, but read sees
// an old value,
func TestWatchRemoveConcur(t *testing.T) {
	const N = 10_000

	ts := makeTstate(t)
	dn := "name/d1"
	err := ts.Mkdir(dn, 0777)
	assert.Equal(t, nil, err)

	fn := dn + "/w"

	ch := make(chan error)
	done := make(chan bool)
	go func() {
		fsl := fslib.MakeFsLibAddr("fsl1", fslib.Named())
		for i := 1; i < N; {
			err := fsl.MakeFile(fn, 0777, np.OWRITE, nil)
			// assert.Equal(t, nil, err)
			if err != nil {
				log.Fatalf("Makefile %v err %v\n", fn, err)
			}
			err = ts.SetRemoveWatch(fn, func(fn string, r error) {
				// log.Printf("watch cb %v err %v\n", i, r)
				ch <- r
			})
			if err == nil {
				r := <-ch
				if r == nil {
					i += 1
				}
			} else {
				// log.Printf("SetRemoveWatch %v err %v\n", i, err)
			}
		}
		done <- true
	}()

	stop := false
	for !stop {
		select {
		case <-done:
			stop = true
		default:
			time.Sleep(1 * time.Millisecond)
			ts.Remove(fn) // remove may fail
		}
	}

	ts.s.Shutdown()
}

func TestConcurFile(t *testing.T) {
	const N = 20
	ts := makeTstate(t)
	ch := make(chan int)
	for i := 0; i < N; i++ {
		go func(i int) {
			for j := 0; j < 1000; j++ {
				fn := "name/f" + strconv.Itoa(i)
				data := []byte(fn)
				err := ts.MakeFile(fn, 0777, np.OWRITE, data)
				assert.Equal(t, nil, err)
				d, err := ts.ReadFile(fn)
				assert.Equal(t, nil, err)
				assert.Equal(t, len(data), len(d))
				err = ts.Remove(fn)
				assert.Equal(t, nil, err)
			}
			ch <- i
		}(i)
	}
	for i := 0; i < N; i++ {
		<-ch
	}
	ts.s.Shutdown()
}

const (
	TODO  = "name/todo"
	DONE  = "name/done"
	NFILE = 1000
)

func (ts *Tstate) initfs() {
	err := ts.Mkdir(TODO, 07000)
	assert.Nil(ts.t, err, "Create done")
	err = ts.Mkdir(DONE, 07000)
	assert.Nil(ts.t, err, "Create todo")
}

// Keep renaming files in the todo directory until we failed to rename
// any file
func (ts *Tstate) testRename(fsl *fslib.FsLib, t string) int {
	ok := true
	i := 0
	for ok {
		ok = false
		sts, err := fsl.ReadDir(TODO)
		assert.Nil(ts.t, err, "ReadDir")
		for _, st := range sts {
			err = fsl.Rename(TODO+"/"+st.Name, DONE+"/"+st.Name+"."+t)
			if err == nil {
				i = i + 1
				ok = true
			} else {
				assert.Contains(ts.t, err.Error(), "file not found")
			}
		}
	}
	return i
}

func (ts *Tstate) checkFs() {
	sts, err := ts.ReadDir(DONE)
	assert.Nil(ts.t, err, "ReadDir")
	assert.Equal(ts.t, NFILE, len(sts), "checkFs")
	files := make(map[int]bool)
	for _, st := range sts {
		n := strings.TrimSuffix(st.Name, filepath.Ext(st.Name))
		n = strings.TrimPrefix(n, "job")
		i, err := strconv.Atoi(n)
		assert.Nil(ts.t, err, "Atoi")
		_, ok := files[i]
		assert.Equal(ts.t, false, ok, "map")
		files[i] = true
	}
	for i := 0; i < NFILE; i++ {
		assert.Equal(ts.t, true, files[i], "checkFs")
	}
}

func TestConcurRename(t *testing.T) {
	const N = 20
	ts := makeTstate(t)
	cont := make(chan bool)
	done := make(chan int)
	ts.initfs()

	// start N threads trying to rename files in todo dir
	for i := 0; i < N; i++ {
		fsl := fslib.MakeFsLibAddr("thread"+strconv.Itoa(i), fslib.Named())
		go func(fsl *fslib.FsLib, t string) {
			n := 0
			for c := true; c; {
				select {
				case c = <-cont:
				default:
					n += ts.testRename(fsl, t)
				}
			}
			done <- n
		}(fsl, strconv.Itoa(i))
	}

	// generate files in the todo dir
	for i := 0; i < NFILE; i++ {
		err := ts.MakeFile(TODO+"/job"+strconv.Itoa(i), 07000, np.OWRITE, []byte{})
		assert.Nil(ts.t, err, "Create job")
	}

	// tell threads we are done with generating files
	n := 0
	for i := 0; i < N; i++ {
		cont <- false
		n += <-done
	}
	assert.Equal(ts.t, NFILE, n, "sum")
	ts.checkFs()
	ts.s.Shutdown()
}

func TestPipeSimple(t *testing.T) {
	ts := makeTstate(t)

	err := ts.MakePipe("name/pipe", 0777)
	assert.Nil(ts.t, err, "MakePipe")

	go func() {
		fsl := fslib.MakeFsLibAddr("reader", fslib.Named())
		fd, err := fsl.Open("name/pipe", np.OREAD)
		assert.Nil(ts.t, err, "Open")
		b, err := fsl.Read(fd, 100)
		assert.Nil(ts.t, err, "Read")
		assert.Equal(ts.t, "hello", string(b))
		err = fsl.Close(fd)
		assert.Nil(ts.t, err, "Close")
	}()
	fd, err := ts.Open("name/pipe", np.OWRITE)
	assert.Nil(ts.t, err, "Open")
	_, err = ts.Write(fd, []byte("hello"))
	assert.Nil(ts.t, err, "Write")
	err = ts.Close(fd)
	assert.Nil(ts.t, err, "Close")

	ts.s.Shutdown()
}

func TestPipeClose(t *testing.T) {
	ts := makeTstate(t)

	err := ts.MakePipe("name/pipe", 0777)
	assert.Nil(ts.t, err, "MakePipe")

	ch := make(chan bool)
	go func(ch chan bool) {
		fsl := fslib.MakeFsLibAddr("reader", fslib.Named())
		fd, err := fsl.Open("name/pipe", np.OREAD)
		assert.Nil(ts.t, err, "Open")
		for true {
			b, err := fsl.Read(fd, 100)
			if err != nil { // writer closed pipe
				break
			}
			assert.Nil(ts.t, err, "Read")
			assert.Equal(ts.t, "hello", string(b))
		}
		err = fsl.Close(fd)
		assert.Nil(ts.t, err, "Close")
		ch <- true
	}(ch)
	fd, err := ts.Open("name/pipe", np.OWRITE)
	assert.Nil(ts.t, err, "Open")
	_, err = ts.Write(fd, []byte("hello"))
	assert.Nil(ts.t, err, "Write")
	err = ts.Close(fd)
	assert.Nil(ts.t, err, "Close")

	<-ch

	ts.s.Shutdown()
}

func TestPipeRemove(t *testing.T) {
	ts := makeTstate(t)

	err := ts.MakePipe("name/pipe", 0777)
	assert.Nil(ts.t, err, "MakePipe")

	ch := make(chan bool)
	go func(ch chan bool) {
		fsl := fslib.MakeFsLibAddr("reader", fslib.Named())
		_, err := fsl.Open("name/pipe", np.OREAD)
		assert.NotNil(ts.t, err, "Open")
		ch <- true
	}(ch)
	time.Sleep(500 * time.Millisecond)
	err = ts.Remove("name/pipe")
	assert.Nil(ts.t, err, "Remove")

	<-ch

	ts.s.Shutdown()
}
