package fslib_test

import (
	"bufio"
	"flag"
	gopath "path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/stretchr/testify/assert"

	db "sigmaos/debug"
	"sigmaos/fslib"
	"sigmaos/named"
	"sigmaos/path"
	"sigmaos/perf"
	"sigmaos/serr"
	"sigmaos/sessp"
	sp "sigmaos/sigmap"
	"sigmaos/stats"
	"sigmaos/test"
)

var pathname string // e.g., --path "name/ux/~local/fslibtest"

func init() {
	flag.StringVar(&pathname, "path", sp.NAMED, "path for file system")
}

func TestInitFs(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)
	sts, err := ts.GetDir(pathname)
	assert.Nil(t, err)
	if pathname == sp.NAMED {
		assert.True(t, fslib.Present(sts, named.InitDir), "initfs")
	} else {
		assert.True(t, len(sts) == 0, "initfs")
	}
	ts.Shutdown()
}

func TestRemoveBasic(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	fn := gopath.Join(pathname, "f")
	d := []byte("hello")
	_, err := ts.PutFile(fn, 0777, sp.OWRITE, d)
	assert.Equal(t, nil, err)

	err = ts.Remove(fn)
	assert.Equal(t, nil, err)

	_, err = ts.Stat(fn)
	assert.NotEqual(t, nil, err)

	ts.Shutdown()
}

func TestCreateTwice(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	fn := gopath.Join(pathname, "f")
	d := []byte("hello")
	_, err := ts.PutFile(fn, 0777, sp.OWRITE, d)
	assert.Equal(t, nil, err)
	_, err = ts.PutFile(fn, 0777, sp.OWRITE, d)
	assert.NotNil(t, err)
	assert.True(t, serr.IsErrExists(err))
	ts.Shutdown()
}

func TestConnect(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	fn := gopath.Join(pathname, "f")
	d := []byte("hello")
	fd, err := ts.Create(fn, 0777, sp.OWRITE)
	assert.Equal(t, nil, err)
	_, err = ts.Write(fd, d)
	assert.Equal(t, nil, err)

	srv, _, err := ts.PathLastSymlink(pathname)
	assert.Nil(t, err)

	err = ts.Disconnect(srv)
	assert.Nil(t, err, "Disconnect")
	time.Sleep(100 * time.Millisecond)
	db.DPrintf(db.ALWAYS, "disconnected")

	_, err = ts.Write(fd, d)
	assert.True(t, serr.IsErrUnreachable(err))

	err = ts.Close(fd)
	assert.True(t, serr.IsErrUnreachable(err))

	fd, err = ts.Open(fn, sp.OREAD)
	assert.True(t, serr.IsErrUnreachable(err))

	ts.Shutdown()
}

func TestRemoveNonExistent(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	fn := gopath.Join(pathname, "f")
	d := []byte("hello")
	_, err := ts.PutFile(fn, 0777, sp.OWRITE, d)
	assert.Equal(t, nil, err)

	err = ts.Remove(gopath.Join(pathname, "this-file-does-not-exist"))
	assert.NotNil(t, err)

	ts.Shutdown()
}

func TestRemovePath(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	d1 := gopath.Join(pathname, "d1")
	err := ts.MkDir(d1, 0777)
	assert.Equal(t, nil, err)
	fn := gopath.Join(d1, "f")
	d := []byte("hello")
	_, err = ts.PutFile(fn, 0777, sp.OWRITE, d)
	assert.Equal(t, nil, err)

	b, err := ts.GetFile(fn)
	assert.Equal(t, "hello", string(b))

	err = ts.Remove(fn)
	assert.Equal(t, nil, err)

	ts.Shutdown()
}

func TestRemoveSymlink(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	d1 := gopath.Join(pathname, "d1")
	db.DPrintf(db.ALWAYS, "path %v", pathname)
	err := ts.MkDir(d1, 0777)
	assert.Nil(t, err, "Mkdir %v", err)
	fn := gopath.Join(d1, "f")

	mnt := sp.MkMountService(ts.NamedAddr())
	err = ts.MkMountSymlink(fn, mnt)
	assert.Nil(t, err, "MkMount: %v", err)

	_, err = ts.GetDir(fn + "/")
	assert.Nil(t, err, "GetDir: %v", err)

	err = ts.Remove(fn)
	assert.Nil(t, err, "RmDir: %v", err)

	ts.Shutdown()
}

func TestRmDirWithSymlink(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	d1 := gopath.Join(pathname, "d1")
	err := ts.MkDir(d1, 0777)
	assert.Nil(t, err, "Mkdir %v", err)
	fn := gopath.Join(d1, "f")

	mnt := sp.MkMountService(ts.NamedAddr())
	err = ts.MkMountSymlink(fn, mnt)
	assert.Nil(t, err, "MkMount: %v", err)

	_, err = ts.GetDir(fn + "/")
	assert.Nil(t, err, "GetDir: %v", err)

	err = ts.RmDir(d1)
	assert.Nil(t, err, "RmDir: %v", err)

	ts.Shutdown()
}

func TestReadOff(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	fn := gopath.Join(pathname, "f")
	d := []byte("hello")
	_, err := ts.PutFile(fn, 0777, sp.OWRITE, d)
	assert.Equal(t, nil, err)

	rdr, err := ts.OpenReader(fn)
	assert.Equal(t, nil, err)

	rdr.Lseek(3)
	b := make([]byte, 10)
	n, err := rdr.Read(b)
	assert.Nil(t, err)
	assert.Equal(t, 2, n)

	ts.Shutdown()
}

func TestRenameBasic(t *testing.T) {
	d1 := gopath.Join(pathname, "d1")
	d2 := gopath.Join(pathname, "d2")
	ts := test.MakeTstatePath(t, pathname)
	err := ts.MkDir(d1, 0777)
	assert.Equal(t, nil, err)
	err = ts.MkDir(d2, 0777)
	assert.Equal(t, nil, err)

	fn := gopath.Join(d1, "f")
	fn1 := gopath.Join(d2, "g")
	d := []byte("hello")
	_, err = ts.PutFile(fn, 0777, sp.OWRITE, d)
	assert.Equal(t, nil, err)

	err = ts.Rename(fn, fn1)
	assert.Equal(t, nil, err)

	b, err := ts.GetFile(fn1)
	assert.Equal(t, "hello", string(b))
	ts.Shutdown()
}

func TestRenameAndRemove(t *testing.T) {
	d1 := gopath.Join(pathname, "d1")
	d2 := gopath.Join(pathname, "d2")
	ts := test.MakeTstatePath(t, pathname)
	err := ts.MkDir(d1, 0777)
	assert.Equal(t, nil, err)
	err = ts.MkDir(d2, 0777)
	assert.Equal(t, nil, err)

	fn := gopath.Join(d1, "f")
	fn1 := gopath.Join(d2, "g")
	d := []byte("hello")
	_, err = ts.PutFile(fn, 0777, sp.OWRITE, d)
	assert.Equal(t, nil, err)

	err = ts.Rename(fn, fn1)
	assert.Equal(t, nil, err)

	b, err := ts.GetFile(fn1)
	assert.Equal(t, nil, err)
	assert.Equal(t, "hello", string(b))

	_, err = ts.Stat(fn1)
	assert.Equal(t, nil, err)

	err = ts.Remove(fn1)
	assert.Equal(t, nil, err)
	ts.Shutdown()
}

func TestNonEmpty(t *testing.T) {
	d1 := gopath.Join(pathname, "d1")
	d2 := gopath.Join(pathname, "d2")

	ts := test.MakeTstatePath(t, pathname)
	err := ts.MkDir(d1, 0777)
	assert.Equal(t, nil, err)
	err = ts.MkDir(d2, 0777)
	assert.Equal(t, nil, err)

	fn := gopath.Join(d1, "f")
	d := []byte("hello")
	_, err = ts.PutFile(fn, 0777, sp.OWRITE, d)
	assert.Equal(t, nil, err)

	err = ts.Remove(d1)
	assert.NotNil(t, err, "Remove")

	err = ts.Rename(d2, d1)
	assert.NotNil(t, err, "Rename")

	ts.Shutdown()
}

func TestSetAppend(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)
	d := []byte("1234")
	fn := gopath.Join(pathname, "f")

	_, err := ts.PutFile(fn, 0777, sp.OWRITE, d)
	assert.Equal(t, nil, err)
	l, err := ts.SetFile(fn, d, sp.OAPPEND, sp.NoOffset)
	assert.Equal(t, nil, err)
	assert.Equal(t, sessp.Tsize(len(d)), l)
	b, err := ts.GetFile(fn)
	assert.Equal(t, nil, err)
	assert.Equal(t, len(d)*2, len(b))
	ts.Shutdown()
}

func TestCopy(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)
	d := []byte("hello")
	src := gopath.Join(pathname, "f")
	dst := gopath.Join(pathname, "g")
	_, err := ts.PutFile(src, 0777, sp.OWRITE, d)
	assert.Equal(t, nil, err)

	err = ts.CopyFile(src, dst)
	assert.Equal(t, nil, err)

	d1, err := ts.GetFile(dst)
	assert.Equal(t, "hello", string(d1))

	ts.Shutdown()
}

func TestDirBasic(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)
	dn := gopath.Join(pathname, "d")
	err := ts.MkDir(dn, 0777)
	assert.Equal(t, nil, err)
	b, err := ts.IsDir(dn)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, b)

	d := []byte("hello")
	_, err = ts.PutFile(gopath.Join(dn, "f"), 0777, sp.OWRITE, d)
	assert.Equal(t, nil, err)

	sts, err := ts.GetDir(dn)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(sts))
	assert.Equal(t, "f", sts[0].Name)

	err = ts.RmDir(dn)
	_, err = ts.Stat(dn)
	assert.NotNil(t, err)

	ts.Shutdown()
}

func TestDirDot(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)
	dn := gopath.Join(pathname, "dir0")
	dot := dn + "/."
	err := ts.MkDir(dn, 0777)
	assert.Equal(t, nil, err)
	b, err := ts.IsDir(dot)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, b)
	err = ts.RmDir(dot)
	assert.NotNil(t, err)
	err = ts.RmDir(dn)
	_, err = ts.Stat(dot)
	assert.NotNil(t, err)
	_, err = ts.Stat(pathname + "/.")
	assert.Nil(t, err, "Couldn't stat %v", err)
	ts.Shutdown()
}

func TestPageDir(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)
	dn := gopath.Join(pathname, "dir")
	err := ts.MkDir(dn, 0777)
	assert.Equal(t, nil, err)
	ts.SetChunkSz(sessp.Tsize(512))
	n := 1000
	names := make([]string, 0)
	for i := 0; i < n; i++ {
		name := strconv.Itoa(i)
		names = append(names, name)
		_, err := ts.PutFile(gopath.Join(dn, name), 0777, sp.OWRITE, []byte(name))
		assert.Equal(t, nil, err)
	}
	sort.SliceStable(names, func(i, j int) bool {
		return names[i] < names[j]
	})
	i := 0
	ts.ProcessDir(dn, func(st *sp.Stat) (bool, error) {
		assert.Equal(t, names[i], st.Name)
		i += 1
		return false, nil

	})
	assert.Equal(t, i, n)
	ts.Shutdown()
}

func dirwriter(t *testing.T, dn string, name string, nds []string, ch chan bool) {
	fsl, err := fslib.MakeFsLibAddr("fslibtest-"+name, nds)
	assert.Nil(t, err)
	stop := false
	for !stop {
		select {
		case stop = <-ch:
		default:
			err := fsl.Remove(gopath.Join(dn, name))
			assert.Nil(t, err)
			_, err = fsl.PutFile(gopath.Join(dn, name), 0777, sp.OWRITE, []byte(name))
			assert.Nil(t, err)
		}
	}
}

// Concurrently scan dir and create/remove entries
func TestDirConcur(t *testing.T) {
	const (
		N     = 1
		NFILE = 3
		NSCAN = 100
	)
	ts := test.MakeTstatePath(t, pathname)
	dn := gopath.Join(pathname, "dir")
	err := ts.MkDir(dn, 0777)
	assert.Equal(t, nil, err)

	for i := 0; i < NFILE; i++ {
		name := strconv.Itoa(i)
		_, err := ts.PutFile(gopath.Join(dn, name), 0777, sp.OWRITE, []byte(name))
		assert.Equal(t, nil, err)
	}

	ch := make(chan bool)
	for i := 0; i < N; i++ {
		go dirwriter(t, dn, strconv.Itoa(i), ts.NamedAddr(), ch)
	}

	for i := 0; i < NSCAN; i++ {
		i := 0
		names := []string{}
		b, err := ts.ProcessDir(dn, func(st *sp.Stat) (bool, error) {
			names = append(names, st.Name)
			i += 1
			return false, nil

		})
		assert.Nil(t, err)
		assert.False(t, b)

		if i < NFILE-N {
			db.DPrintf(db.ALWAYS, "names %v", names)
		}

		assert.True(t, i >= NFILE-N)

		uniq := make(map[string]bool)
		for _, n := range names {
			if _, ok := uniq[n]; ok {
				assert.True(t, n == strconv.Itoa(NFILE-1))
			}
			uniq[n] = true
		}
	}

	for i := 0; i < N; i++ {
		ch <- true
	}

	ts.Shutdown()
}

func readWrite(t *testing.T, fsl *fslib.FsLib, cnt string) bool {
	fd, err := fsl.Open(cnt, sp.ORDWR)
	assert.Nil(t, err)

	defer fsl.Close(fd)

	b, err := fsl.ReadV(fd, 1000)
	if err != nil && serr.IsErrVersion(err) {
		return true
	}
	assert.Nil(t, err)
	n, err := strconv.Atoi(string(b))
	assert.Nil(t, err)

	n += 1

	err = fsl.Seek(fd, 0)
	assert.Nil(t, err)

	b = []byte(strconv.Itoa(n))
	_, err = fsl.WriteV(fd, b)
	if err != nil && serr.IsErrVersion(err) {
		return true
	}
	assert.Nil(t, err)

	return false
}

func TestCounter(t *testing.T) {
	const N = 10

	ts := test.MakeTstatePath(t, pathname)
	cnt := gopath.Join(pathname, "cnt")
	b := []byte(strconv.Itoa(0))
	_, err := ts.PutFile(cnt, 0777|sp.DMTMP, sp.OWRITE, b)
	assert.Equal(t, nil, err)

	ch := make(chan int)

	for i := 0; i < N; i++ {
		go func(i int) {
			ntrial := 0
			for {
				ntrial += 1
				if readWrite(t, ts.FsLib, cnt) {
					continue
				}
				break
			}
			// log.Printf("%d: tries %v\n", i, ntrial)
			ch <- i
		}(i)
	}
	for i := 0; i < N; i++ {
		<-ch
	}
	b, err = ts.GetFile(cnt)
	assert.Equal(t, nil, err)
	n, err := strconv.Atoi(string(b))
	assert.Equal(t, nil, err)

	assert.Equal(t, N, n)

	ts.Shutdown()
}

func TestWatchCreate(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	fn := gopath.Join(pathname, "w")
	ch := make(chan bool)
	fd, err := ts.OpenWatch(fn, sp.OREAD, func(string, error) {
		ch <- true
	})
	assert.NotEqual(t, nil, err)
	assert.Equal(t, -1, fd, err)
	assert.True(t, serr.IsErrNotfound(err))

	// give Watch goroutine to start
	time.Sleep(100 * time.Millisecond)

	_, err = ts.PutFile(fn, 0777, sp.OWRITE, nil)
	assert.Equal(t, nil, err)

	<-ch

	ts.Shutdown()
}

func TestWatchRemoveOne(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	fn := gopath.Join(pathname, "w")
	_, err := ts.PutFile(fn, 0777, sp.OWRITE, nil)
	assert.Equal(t, nil, err)

	ch := make(chan bool)
	err = ts.SetRemoveWatch(fn, func(path string, err error) {
		assert.Equal(t, nil, err, path)
		ch <- true
	})
	assert.Equal(t, nil, err)

	// give Watch goroutine to start
	time.Sleep(100 * time.Millisecond)

	err = ts.Remove(fn)
	assert.Equal(t, nil, err)

	<-ch

	ts.Shutdown()
}

func TestWatchDir(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	fn := gopath.Join(pathname, "d1")
	err := ts.MkDir(fn, 0777)
	assert.Equal(t, nil, err)

	_, rdr, err := ts.ReadDir(fn)
	assert.Equal(t, nil, err)
	ch := make(chan bool)
	err = ts.SetDirWatch(rdr.Fid(), fn, func(path string, err error) {
		assert.Equal(t, nil, err, path)
		ch <- true
	})
	assert.Equal(t, nil, err)

	// give Watch goroutine to start
	time.Sleep(100 * time.Millisecond)

	_, err = ts.PutFile(gopath.Join(fn, "x"), 0777, sp.OWRITE, nil)
	assert.Equal(t, nil, err)

	<-ch

	ts.Shutdown()
}

func TestCreateExcl1(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)
	ch := make(chan int)

	fn := gopath.Join(pathname, "exclusive")
	_, err := ts.PutFile(fn, 0777|sp.DMTMP, sp.OWRITE|sp.OCEXEC, []byte{})
	assert.Nil(t, err)
	fsl, err := fslib.MakeFsLibAddr("fslibtest0", ts.NamedAddr())
	assert.Nil(t, err)
	go func() {
		_, err := fsl.PutFile(fn, 0777|sp.DMTMP, sp.OWRITE|sp.OWATCH, []byte{})
		assert.Nil(t, err, "Putfile")
		ch <- 0
	}()
	time.Sleep(time.Second * 2)
	err = ts.Remove(fn)
	assert.Nil(t, err, "Remove")
	go func() {
		time.Sleep(2 * time.Second)
		ch <- 1
	}()
	i := <-ch
	assert.Equal(t, 0, i)

	ts.Shutdown()
}

func TestCreateExclN(t *testing.T) {
	const N = 20

	ts := test.MakeTstatePath(t, pathname)
	ch := make(chan int)
	fn := gopath.Join(pathname, "exclusive")
	acquired := false
	for i := 0; i < N; i++ {
		go func(i int) {
			fsl, err := fslib.MakeFsLibAddr("fslibtest"+strconv.Itoa(i), ts.NamedAddr())
			assert.Nil(t, err)
			_, err = fsl.PutFile(fn, 0777|sp.DMTMP, sp.OWRITE|sp.OWATCH, []byte{})
			assert.Equal(t, nil, err)
			assert.Equal(t, false, acquired)
			acquired = true
			ch <- i
		}(i)
	}
	for i := 0; i < N; i++ {
		<-ch
		acquired = false
		err := ts.Remove(fn)
		assert.Equal(t, nil, err)
	}
	ts.Shutdown()
}

func TestCreateExclAfterDisconnect(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	fn := gopath.Join(pathname, "create-conn-close-test")

	fsl1, err := fslib.MakeFsLibAddr("fslibtest-1", ts.NamedAddr())
	assert.Nil(t, err)
	_, err = ts.PutFile(fn, 0777|sp.DMTMP, sp.OWRITE|sp.OWATCH, []byte{})
	assert.Nil(t, err, "Create 1")

	go func() {
		// Should wait
		_, err := fsl1.PutFile(fn, 0777|sp.DMTMP, sp.OWRITE|sp.OWATCH, []byte{})
		assert.NotNil(t, err, "Create 2")
	}()

	time.Sleep(500 * time.Millisecond)

	// Kill fsl1's connection
	srv, _, err := ts.PathLastSymlink(pathname)
	assert.Nil(t, err)

	db.DPrintf(db.TEST, "Disconnect fsl")
	err = fsl1.Disconnect(srv)
	assert.Nil(t, err, "Disconnect")

	// Remove the ephemeral file
	ts.Remove(fn)
	assert.Equal(t, nil, err)

	// Try to create again (should succeed)
	_, err = ts.PutFile(fn, 0777|sp.DMTMP, sp.OWRITE|sp.OWATCH, []byte{})
	assert.Nil(t, err, "Create 3")

	ts.Shutdown()
}

func TestWatchRemoveConcur(t *testing.T) {
	const N = 5_000

	ts := test.MakeTstatePath(t, pathname)
	dn := gopath.Join(pathname, "d1")
	err := ts.MkDir(dn, 0777)
	assert.Equal(t, nil, err)

	fn := gopath.Join(dn, "w")

	ch := make(chan error)
	done := make(chan bool)
	go func() {
		fsl, err := fslib.MakeFsLibAddr("fsl1", ts.NamedAddr())
		assert.Nil(t, err)
		for i := 1; i < N; {
			_, err := fsl.PutFile(fn, 0777, sp.OWRITE, nil)
			assert.Equal(t, nil, err)
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

	ts.Shutdown()
}

// Concurrently remove & watch, but watch may be set after remove.
func TestWatchRemoveConcurAsynchWatchSet(t *testing.T) {
	const N = 10_000

	ts := test.MakeTstatePath(t, pathname)
	dn := gopath.Join(pathname, "d1")
	err := ts.MkDir(dn, 0777)
	assert.Equal(t, nil, err)

	ch := make(chan error)
	done := make(chan bool)
	fsl, err := fslib.MakeFsLibAddr("fsl1", ts.NamedAddr())
	assert.Nil(t, err)
	for i := 0; i < N; i++ {
		fn := gopath.Join(dn, strconv.Itoa(i))
		_, err := fsl.PutFile(fn, 0777, sp.OWRITE, nil)
		assert.Nil(t, err, "Err putfile: %v", err)
	}
	for i := 0; i < N; i++ {
		fn := gopath.Join(dn, strconv.Itoa(i))
		go func(fn string) {
			err := ts.SetRemoveWatch(fn, func(fn string, r error) {
				// log.Printf("watch cb %v err %v\n", i, r)
				ch <- r
			})
			// Either no error, or remove already happened.
			assert.True(ts.T, err == nil || serr.IsErrNotfound(err), "Unexpected RemoveWatch error: %v", err)
			done <- true
		}(fn)
		go func(fn string) {
			err := ts.Remove(fn)
			assert.Nil(t, err, "Unexpected remove error: %v", err)
		}(fn)
	}
	for i := 0; i < N; i++ {
		<-done
	}
	ts.Shutdown()
}

func TestConcurFile(t *testing.T) {
	const N = 20
	ts := test.MakeTstatePath(t, pathname)
	ch := make(chan int)
	for i := 0; i < N; i++ {
		go func(i int) {
			for j := 0; j < 1000; j++ {
				fn := gopath.Join(pathname, "f"+strconv.Itoa(i))
				data := []byte(fn)
				_, err := ts.PutFile(fn, 0777, sp.OWRITE, data)
				assert.Equal(t, nil, err)
				d, err := ts.GetFile(fn)
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
	ts.Shutdown()
}

const (
	NFILE = 1000
)

func initfs(ts *test.Tstate, TODO, DONE string) {
	err := ts.MkDir(TODO, 07777)
	assert.Nil(ts.T, err, "Create done")
	err = ts.MkDir(DONE, 07777)
	assert.Nil(ts.T, err, "Create todo")
}

// Keep renaming files in the todo directory until we failed to rename
// any file
func testRename(ts *test.Tstate, fsl *fslib.FsLib, t string, TODO, DONE string) int {
	ok := true
	i := 0
	for ok {
		ok = false
		sts, err := fsl.GetDir(TODO)
		assert.Nil(ts.T, err, "GetDir")
		for _, st := range sts {
			err = fsl.Rename(gopath.Join(TODO, st.Name), gopath.Join(DONE, st.Name+"."+t))
			if err == nil {
				i = i + 1
				ok = true
			} else {
				assert.True(ts.T, serr.IsErrNotfound(err))
			}
		}
	}
	return i
}

func checkFs(ts *test.Tstate, DONE string) {
	sts, err := ts.GetDir(DONE)
	assert.Nil(ts.T, err, "GetDir")
	assert.Equal(ts.T, NFILE, len(sts), "checkFs")
	files := make(map[int]bool)
	for _, st := range sts {
		n := strings.TrimSuffix(st.Name, filepath.Ext(st.Name))
		n = strings.TrimPrefix(n, "job")
		i, err := strconv.Atoi(n)
		assert.Nil(ts.T, err, "Atoi")
		_, ok := files[i]
		assert.Equal(ts.T, false, ok, "map")
		files[i] = true
	}
	for i := 0; i < NFILE; i++ {
		assert.Equal(ts.T, true, files[i], "checkFs")
	}
}

func TestConcurRename(t *testing.T) {
	const N = 20
	ts := test.MakeTstatePath(t, pathname)
	cont := make(chan bool)
	done := make(chan int)
	TODO := gopath.Join(pathname, "todo")
	DONE := gopath.Join(pathname, "done")

	initfs(ts, TODO, DONE)

	// start N threads trying to rename files in todo dir
	for i := 0; i < N; i++ {
		fsl, err := fslib.MakeFsLibAddr("thread"+strconv.Itoa(i), ts.NamedAddr())
		assert.Nil(t, err)
		go func(fsl *fslib.FsLib, t string) {
			n := 0
			for c := true; c; {
				select {
				case c = <-cont:
				default:
					n += testRename(ts, fsl, t, TODO, DONE)
				}
			}
			done <- n
		}(fsl, strconv.Itoa(i))
	}

	// generate files in the todo dir
	for i := 0; i < NFILE; i++ {
		_, err := ts.PutFile(gopath.Join(TODO, "job"+strconv.Itoa(i)), 07000, sp.OWRITE, []byte{})
		assert.Nil(ts.T, err, "Create job")
	}

	// tell threads we are done with generating files
	n := 0
	for i := 0; i < N; i++ {
		cont <- false
		n += <-done
	}
	assert.Equal(ts.T, NFILE, n, "sum")
	checkFs(ts, DONE)
	ts.Shutdown()
}

func TestPipeBasic(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	pipe := gopath.Join(pathname, "pipe")
	err := ts.MakePipe(pipe, 0777)
	assert.Nil(ts.T, err, "MakePipe")

	ch := make(chan bool)
	go func() {
		fsl, err := fslib.MakeFsLibAddr("reader", ts.NamedAddr())
		assert.Nil(t, err)
		fd, err := fsl.Open(pipe, sp.OREAD)
		assert.Nil(ts.T, err, "Open")
		b, err := fsl.Read(fd, 100)
		assert.Nil(ts.T, err, "Read")
		assert.Equal(ts.T, "hello", string(b))
		err = fsl.Close(fd)
		assert.Nil(ts.T, err, "Close")
		ch <- true
	}()
	fd, err := ts.Open(pipe, sp.OWRITE)
	assert.Nil(ts.T, err, "Open")
	_, err = ts.Write(fd, []byte("hello"))
	assert.Nil(ts.T, err, "Write")
	err = ts.Close(fd)
	assert.Nil(ts.T, err, "Close")

	<-ch

	ts.Remove(pipe)

	ts.Shutdown()
}

func TestPipeClose(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	pipe := gopath.Join(pathname, "pipe")
	err := ts.MakePipe(pipe, 0777)
	assert.Nil(ts.T, err, "MakePipe")

	ch := make(chan bool)
	go func(ch chan bool) {
		fsl, err := fslib.MakeFsLibAddr("reader", ts.NamedAddr())
		assert.Nil(t, err)
		fd, err := fsl.Open(pipe, sp.OREAD)
		assert.Nil(ts.T, err, "Open")
		for true {
			b, err := fsl.Read(fd, 100)
			if err != nil { // writer closed pipe
				break
			}
			assert.Nil(ts.T, err, "Read")
			assert.Equal(ts.T, "hello", string(b))
		}
		err = fsl.Close(fd)
		assert.Nil(ts.T, err, "Close: %v", err)
		ch <- true
	}(ch)
	fd, err := ts.Open(pipe, sp.OWRITE)
	assert.Nil(ts.T, err, "Open")
	_, err = ts.Write(fd, []byte("hello"))
	assert.Nil(ts.T, err, "Write")
	err = ts.Close(fd)
	assert.Nil(ts.T, err, "Close")

	<-ch

	ts.Remove(pipe)

	ts.Shutdown()
}

func TestPipeRemove(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)
	pipe := gopath.Join(pathname, "pipe")

	err := ts.MakePipe(pipe, 0777)
	assert.Nil(ts.T, err, "MakePipe")

	ch := make(chan bool)
	go func(ch chan bool) {
		fsl, err := fslib.MakeFsLibAddr("reader", ts.NamedAddr())
		assert.Nil(t, err)
		_, err = fsl.Open(pipe, sp.OREAD)
		assert.NotNil(ts.T, err, "Open")
		ch <- true
	}(ch)
	time.Sleep(500 * time.Millisecond)
	err = ts.Remove(pipe)
	assert.Nil(ts.T, err, "Remove")

	<-ch

	ts.Shutdown()
}

func TestPipeCrash0(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)
	pipe := gopath.Join(pathname, "pipe")
	err := ts.MakePipe(pipe, 0777)
	assert.Nil(ts.T, err, "MakePipe")

	go func() {
		fsl, err := fslib.MakeFsLibAddr("writer", ts.NamedAddr())
		assert.Nil(t, err)
		_, err = fsl.Open(pipe, sp.OWRITE)
		assert.Nil(ts.T, err, "Open")
		time.Sleep(200 * time.Millisecond)
		// simulate thread crashing
		srv, _, err := ts.PathLastSymlink(pathname)
		assert.Nil(t, err)
		err = fsl.Disconnect(srv)
		assert.Nil(ts.T, err, "Disconnect")

	}()
	fd, err := ts.Open(pipe, sp.OREAD)
	assert.Nil(ts.T, err, "Open")
	_, err = ts.Read(fd, 100)
	assert.NotNil(ts.T, err, "read")

	ts.Remove(pipe)
	ts.Shutdown()
}

func TestPipeCrash1(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)
	pipe := gopath.Join(pathname, "pipe")
	err := ts.MakePipe(pipe, 0777)
	assert.Nil(ts.T, err, "MakePipe")

	fsl1, err := fslib.MakeFsLibAddr("w1", ts.NamedAddr())
	assert.Nil(t, err)
	go func() {
		// blocks
		_, err := fsl1.Open(pipe, sp.OWRITE)
		assert.NotNil(ts.T, err, "Open")
	}()

	time.Sleep(200 * time.Millisecond)

	// simulate crash of w1
	srv, _, err := ts.PathLastSymlink(pathname)
	assert.Nil(t, err)
	err = fsl1.Disconnect(srv)
	assert.Nil(ts.T, err, "Disconnect")

	time.Sleep(2 * sp.Conf.Session.TIMEOUT)

	// start up second write to pipe
	go func() {
		fsl2, err := fslib.MakeFsLibAddr("w2", ts.NamedAddr())
		assert.Nil(t, err)
		// the pipe has been closed for writing due to crash;
		// this open should fail.
		_, err = fsl2.Open(pipe, sp.OWRITE)
		assert.NotNil(ts.T, err, "Open")
	}()

	time.Sleep(200 * time.Millisecond)

	fd, err := ts.Open(pipe, sp.OREAD)
	assert.Nil(ts.T, err, "Open")
	_, err = ts.Read(fd, 100)
	assert.NotNil(ts.T, err, "read")

	ts.Remove(pipe)
	ts.Shutdown()
}

func TestSymlinkPath(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	dn := gopath.Join(pathname, "d")
	err := ts.MkDir(dn, 0777)
	assert.Nil(ts.T, err, "dir")

	err = ts.Symlink([]byte(pathname), gopath.Join(pathname, "namedself"), 0777|sp.DMTMP)
	assert.Nil(ts.T, err, "Symlink")

	sts, err := ts.GetDir(gopath.Join(pathname, "namedself") + "/")
	assert.Equal(t, nil, err)
	assert.True(t, fslib.Present(sts, path.Path{"d", "namedself"}), "dir")

	ts.Shutdown()
}

func mkMount(t *testing.T, ts *test.Tstate, path string) sp.Tmount {
	mnt, left, err := ts.CopyMount(pathname)
	assert.Nil(t, err)
	mnt.SetTree(left)
	return mnt
}

func TestMount(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	dn := gopath.Join(pathname, "d")
	err := ts.MkDir(dn, 0777)
	assert.Nil(ts.T, err, "dir")

	err = ts.MountService(gopath.Join(pathname, "namedself"), mkMount(t, ts, pathname))
	assert.Nil(ts.T, err, "MountService")
	sts, err := ts.GetDir(gopath.Join(pathname, "namedself") + "/")
	assert.Equal(t, nil, err)
	assert.True(t, fslib.Present(sts, path.Path{"d", "namedself"}), "dir")

	ts.Shutdown()
}

func TestUnionDir(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	dn := gopath.Join(pathname, "d")
	err := ts.MkDir(dn, 0777)
	assert.Nil(ts.T, err, "dir")

	err = ts.MountService(gopath.Join(pathname, "d/namedself0"), mkMount(t, ts, pathname))
	assert.Nil(ts.T, err, "MountService")

	err = ts.MountService(gopath.Join(pathname, "d/namedself1"), sp.MkMountServer(":2222"))
	assert.Nil(ts.T, err, "MountService")

	sts, err := ts.GetDir(gopath.Join(pathname, "d/~any") + "/")
	assert.Equal(t, nil, err)
	assert.True(t, fslib.Present(sts, path.Path{"d"}), "dir")

	sts, err = ts.GetDir(gopath.Join(pathname, "d/~any/d") + "/")
	assert.Equal(t, nil, err)
	assert.True(t, fslib.Present(sts, path.Path{"namedself0", "namedself1"}), "dir")

	// // XXX these will fail since named runs with a different IP address on this machine
	sts, err = ts.GetDir(gopath.Join(pathname, "d/~local") + "/")
	assert.Equal(t, nil, err)
	assert.True(t, fslib.Present(sts, path.Path{"d"}), "dir")

	pn, err := ts.ResolveUnions(gopath.Join(pathname, "d/~local"))
	assert.Equal(t, nil, err)
	sts, err = ts.GetDir(pn)
	assert.Nil(t, err)
	assert.True(t, fslib.Present(sts, path.Path{"d"}), "dir")

	ts.Shutdown()
}

func TestUnionRoot(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	err := ts.MountService(gopath.Join(pathname, "namedself0"), mkMount(t, ts, pathname))
	assert.Nil(ts.T, err, "MountService")
	err = ts.MountService(gopath.Join(pathname, "namedself1"), sp.MkMountServer("xxx"))
	assert.Nil(ts.T, err, "MountService")

	sts, err := ts.GetDir(gopath.Join(pathname, "~any") + "/")
	assert.Equal(t, nil, err)
	assert.True(t, fslib.Present(sts, path.Path{"namedself0", "namedself1"}), "dir")

	ts.Shutdown()
}

func TestUnionSymlinkRead(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	mnt := mkMount(t, ts, pathname)
	err := ts.MountService(gopath.Join(pathname, "namedself0"), mnt)
	assert.Nil(ts.T, err, "MountService")

	dn := gopath.Join(pathname, "d")
	err = ts.MkDir(dn, 0777)
	assert.Nil(ts.T, err, "dir")

	err = ts.MountService(gopath.Join(pathname, "d/namedself1"), mnt)
	assert.Nil(ts.T, err, "MountService")

	sts, err := ts.GetDir(gopath.Join(pathname, "~any/d/namedself1") + "/")
	assert.Equal(t, nil, err)
	assert.True(t, fslib.Present(sts, path.Path{"d", "namedself0"}), "root wrong")

	sts, err = ts.GetDir(gopath.Join(pathname, "~any/d/namedself1/d") + "/")
	assert.Equal(t, nil, err)
	assert.True(t, fslib.Present(sts, path.Path{"namedself1"}), "d wrong")

	ts.Shutdown()
}

func TestUnionSymlinkPut(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	err := ts.MountService(gopath.Join(pathname, "namedself0"), mkMount(t, ts, pathname))
	assert.Nil(ts.T, err, "MountService")

	b := []byte("hello")
	fn := gopath.Join(pathname, "~any/namedself0/f")
	_, err = ts.PutFile(fn, 0777, sp.OWRITE, b)
	assert.Equal(t, nil, err)

	fn1 := gopath.Join(pathname, "~any/namedself0/g")
	_, err = ts.PutFile(fn1, 0777, sp.OWRITE, b)
	assert.Equal(t, nil, err)

	sts, err := ts.GetDir(gopath.Join(pathname, "~any/namedself0") + "/")
	assert.Equal(t, nil, err)
	assert.True(t, fslib.Present(sts, path.Path{"f", "g"}), "root wrong")

	d, err := ts.GetFile(gopath.Join(pathname, "~any/namedself0/f"))
	assert.Nil(ts.T, err, "GetFile")
	assert.Equal(ts.T, b, d, "GetFile")

	d, err = ts.GetFile(gopath.Join(pathname, "~any/namedself0/g"))
	assert.Nil(ts.T, err, "GetFile")
	assert.Equal(ts.T, b, d, "GetFile")

	ts.Shutdown()
}

func TestSetFileSymlink(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	fn := gopath.Join(pathname, "f")
	d := []byte("hello")
	_, err := ts.PutFile(fn, 0777, sp.OWRITE, d)
	assert.Equal(t, nil, err)

	err = ts.MountService(gopath.Join(pathname, "namedself0"), mkMount(t, ts, pathname))
	assert.Nil(ts.T, err, "MountService")

	st := stats.StatInfo{}
	err = ts.GetFileJson(gopath.Join("name", sp.STATSD), &st)
	assert.Nil(t, err, "statsd")
	nwalk := st.Nwalk

	d = []byte("byebye")
	n, err := ts.SetFile(gopath.Join(pathname, "namedself0/f"), d, sp.OWRITE, 0)
	assert.Nil(ts.T, err, "SetFile: %v", err)
	assert.Equal(ts.T, sessp.Tsize(len(d)), n, "SetFile")

	err = ts.GetFileJson(gopath.Join(pathname, sp.STATSD), &st)
	assert.Nil(t, err, "statsd")

	assert.NotEqual(ts.T, nwalk, st.Nwalk, "setfile")
	nwalk = st.Nwalk

	b, err := ts.GetFile(gopath.Join(pathname, "namedself0/f"))
	assert.Nil(ts.T, err, "GetFile")
	assert.Equal(ts.T, d, b, "GetFile")

	err = ts.GetFileJson(gopath.Join(pathname, sp.STATSD), &st)
	assert.Nil(t, err, "statsd")

	assert.Equal(ts.T, nwalk, st.Nwalk, "getfile")

	ts.Shutdown()
}

func TestOpenRemoveRead(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	fn := gopath.Join(pathname, "f")
	d := []byte("hello")
	_, err := ts.PutFile(fn, 0777, sp.OWRITE, d)
	assert.Equal(t, nil, err)

	rdr, err := ts.OpenReader(fn)
	assert.Equal(t, nil, err)

	err = ts.Remove(fn)
	assert.Equal(t, nil, err)

	b, err := rdr.GetData()
	assert.Equal(t, nil, err)
	assert.Equal(t, d, b, "data")

	rdr.Close()

	_, err = ts.Stat(fn)
	assert.NotNil(t, err, "stat")

	ts.Shutdown()
}

func TestFslibExit(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)

	dot := pathname + "/."

	// connect
	_, err := ts.Stat(dot)
	assert.Nil(t, err)

	// close
	err = ts.Exit()
	assert.Nil(t, err)

	_, err = ts.Stat(dot)
	assert.NotNil(t, err)
	assert.True(t, serr.IsErrUnreachable(err))

	ts.Shutdown()
}

const (
	KBYTE      = 1 << 10
	NRUNS      = 3
	SYNCFILESZ = 100 * KBYTE
	FILESZ     = 100 * sp.MBYTE
	WRITESZ    = 4096
)

func measure(p *perf.Perf, msg string, f func() sp.Tlength) sp.Tlength {
	totStart := time.Now()
	tot := sp.Tlength(0)
	for i := 0; i < NRUNS; i++ {
		start := time.Now()
		sz := f()
		tot += sz
		p.TptTick(float64(sz))
		ms := time.Since(start).Milliseconds()
		db.DPrintf(db.TEST, "%v: %s took %vms (%s)", msg, humanize.Bytes(uint64(sz)), ms, test.TputStr(sz, ms))
	}
	ms := time.Since(totStart).Milliseconds()
	db.DPrintf(db.ALWAYS, "Average %v: %s took %vms (%s)", msg, humanize.Bytes(uint64(tot)), ms, test.TputStr(tot, ms))
	return tot
}

func measuredir(msg string, nruns int, f func() int) {
	tot := float64(0)
	n := 0
	for i := 0; i < nruns; i++ {
		start := time.Now()
		n += f()
		ms := time.Since(start).Milliseconds()
		tot += float64(ms)
	}
	s := tot / 1000
	db.DPrintf(db.TEST, "%v: %d entries took %vms (%.1f file/s)", msg, n, tot, float64(n)/s)
}

type Thow uint8

const (
	HSYNC Thow = iota + 1
	HBUF
	HASYNC
)

func mkFile(t *testing.T, fsl *fslib.FsLib, fn string, how Thow, buf []byte, sz sp.Tlength) sp.Tlength {
	switch how {
	case HSYNC:
		w, err := fsl.CreateWriter(fn, 0777, sp.OWRITE)
		assert.Nil(t, err, "Error Create writer: %v", err)
		err = test.Writer(t, w, buf, sz)
		assert.Nil(t, err)
		err = w.Close()
		assert.Nil(t, err)
	case HBUF:
		w, err := fsl.CreateWriter(fn, 0777, sp.OWRITE)
		assert.Nil(t, err, "Error Create writer: %v", err)
		bw := bufio.NewWriterSize(w, sp.BUFSZ)
		err = test.Writer(t, bw, buf, sz)
		assert.Nil(t, err)
		err = bw.Flush()
		assert.Nil(t, err)
		err = w.Close()
		assert.Nil(t, err)
	case HASYNC:
		w, err := fsl.CreateAsyncWriter(fn, 0777, sp.OWRITE)
		assert.Nil(t, err, "Error Create writer: %v", err)
		err = test.Writer(t, w, buf, sz)
		assert.Nil(t, err)
		err = w.Close()
		assert.Nil(t, err)
	}
	st, err := fsl.Stat(fn)
	assert.Nil(t, err)
	assert.Equal(t, sp.Tlength(sz), st.Tlength(), "stat")
	return sz
}

func TestWriteFilePerfSingle(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)
	fn := gopath.Join(pathname, "f")
	buf := test.MkBuf(WRITESZ)
	// Remove just in case it was left over from a previous run.
	ts.Remove(fn)
	p1, err := perf.MakePerfMulti("TEST", "writer")
	assert.Nil(t, err)
	defer p1.Done()
	measure(p1, "writer", func() sp.Tlength {
		sz := mkFile(t, ts.FsLib, fn, HSYNC, buf, SYNCFILESZ)
		err := ts.Remove(fn)
		assert.Nil(t, err)
		return sz
	})
	p2, err := perf.MakePerfMulti("TEST", "bufwriter")
	assert.Nil(t, err)
	defer p2.Done()
	measure(p2, "bufwriter", func() sp.Tlength {
		sz := mkFile(t, ts.FsLib, fn, HBUF, buf, FILESZ)
		err := ts.Remove(fn)
		assert.Nil(t, err)
		return sz
	})
	p3, err := perf.MakePerfMulti("TEST", "abufwriter")
	assert.Nil(t, err)
	defer p3.Done()
	measure(p3, "abufwriter", func() sp.Tlength {
		sz := mkFile(t, ts.FsLib, fn, HASYNC, buf, FILESZ)
		err := ts.Remove(fn)
		assert.Nil(t, err)
		return sz
	})
	ts.Shutdown()
}

func TestWriteFilePerfMultiClient(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)
	N_CLI := 10
	buf := test.MkBuf(WRITESZ)
	done := make(chan sp.Tlength)
	fns := make([]string, 0, N_CLI)
	fsls := make([]*fslib.FsLib, 0, N_CLI)
	for i := 0; i < N_CLI; i++ {
		fns = append(fns, gopath.Join(pathname, "f"+strconv.Itoa(i)))
		fsl, err := fslib.MakeFsLibAddr("test"+strconv.Itoa(i), ts.NamedAddr())
		assert.Nil(t, err)
		fsls = append(fsls, fsl)
	}
	// Remove just in case it was left over from a previous run.
	for _, fn := range fns {
		ts.Remove(fn)
	}
	p1, err := perf.MakePerfMulti("TEST", "writer")
	assert.Nil(t, err)
	defer p1.Done()
	start := time.Now()
	for i := range fns {
		go func(i int) {
			n := measure(p1, "writer", func() sp.Tlength {
				sz := mkFile(t, fsls[i], fns[i], HSYNC, buf, SYNCFILESZ)
				err := ts.Remove(fns[i])
				assert.Nil(t, err, "Remove err %v", err)
				return sz
			})
			done <- n
		}(i)
	}
	n := sp.Tlength(0)
	for _ = range fns {
		n += <-done
	}
	ms := time.Since(start).Milliseconds()
	db.DPrintf(db.ALWAYS, "Total tpt writer: %s took %vms (%s)", humanize.Bytes(uint64(n)), ms, test.TputStr(n, ms))
	p2, err := perf.MakePerfMulti("TEST", "bufwriter")
	assert.Nil(t, err)
	defer p2.Done()
	start = time.Now()
	for i := range fns {
		go func(i int) {
			n := measure(p2, "bufwriter", func() sp.Tlength {
				sz := mkFile(t, fsls[i], fns[i], HBUF, buf, FILESZ)
				err := ts.Remove(fns[i])
				assert.Nil(t, err, "Remove err %v", err)
				return sz
			})
			done <- n
		}(i)
	}
	n = 0
	for _ = range fns {
		n += <-done
	}
	ms = time.Since(start).Milliseconds()
	db.DPrintf(db.ALWAYS, "Total tpt bufwriter: %s took %vms (%s)", humanize.Bytes(uint64(n)), ms, test.TputStr(n, ms))
	p3, err := perf.MakePerfMulti("TEST", "abufwriter")
	assert.Nil(t, err)
	defer p3.Done()
	start = time.Now()
	for i := range fns {
		go func(i int) {
			n := measure(p3, "abufwriter", func() sp.Tlength {
				sz := mkFile(t, fsls[i], fns[i], HASYNC, buf, FILESZ)
				err := ts.Remove(fns[i])
				assert.Nil(t, err, "Remove err %v", err)
				return sz
			})
			done <- n
		}(i)
	}
	n = 0
	for _ = range fns {
		n += <-done
	}
	ms = time.Since(start).Milliseconds()
	db.DPrintf(db.ALWAYS, "Total tpt bufwriter: %s took %vms (%s)", humanize.Bytes(uint64(n)), ms, test.TputStr(n, ms))
	ts.Shutdown()
}

func TestReadFilePerfSingle(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)
	fn := gopath.Join(pathname, "f")
	buf := test.MkBuf(WRITESZ)
	// Remove just in case it was left over from a previous run.
	ts.Remove(fn)
	sz := mkFile(t, ts.FsLib, fn, HBUF, buf, SYNCFILESZ)
	p1, r := perf.MakePerfMulti("TEST", "reader")
	assert.Nil(t, r)
	defer p1.Done()
	measure(p1, "reader", func() sp.Tlength {
		r, err := ts.OpenReader(fn)
		assert.Nil(t, err)
		n, err := test.Reader(t, r, buf, sz)
		assert.Nil(t, err)
		r.Close()
		return n
	})
	err := ts.Remove(fn)
	assert.Nil(t, err)
	p2, err := perf.MakePerfMulti("TEST", "bufreader")
	assert.Nil(t, err)
	defer p2.Done()
	sz = mkFile(t, ts.FsLib, fn, HBUF, buf, FILESZ)
	measure(p2, "bufreader", func() sp.Tlength {
		r, err := ts.OpenReader(fn)
		assert.Nil(t, err)
		br := bufio.NewReaderSize(r, sp.BUFSZ)
		n, err := test.Reader(t, br, buf, sz)
		assert.Nil(t, err)
		r.Close()
		return n
	})
	p3, err := perf.MakePerfMulti("TEST", "abufreader")
	assert.Nil(t, err)
	defer p3.Done()
	measure(p3, "readahead", func() sp.Tlength {
		r, err := ts.OpenAsyncReader(fn, 0)
		assert.Nil(t, err)
		n, err := test.Reader(t, r, buf, sz)
		assert.Nil(t, err)
		r.Close()
		return n
	})
	err = ts.Remove(fn)
	assert.Nil(t, err)
	ts.Shutdown()
}

func TestReadFilePerfMultiClient(t *testing.T) {
	ts := test.MakeTstatePath(t, pathname)
	N_CLI := 10
	buf := test.MkBuf(WRITESZ)
	done := make(chan sp.Tlength)
	fns := make([]string, 0, N_CLI)
	fsls := make([]*fslib.FsLib, 0, N_CLI)
	for i := 0; i < N_CLI; i++ {
		fns = append(fns, gopath.Join(pathname, "f"+strconv.Itoa(i)))
		fsl, err := fslib.MakeFsLibAddr("test"+strconv.Itoa(i), ts.NamedAddr())
		assert.Nil(t, err)
		fsls = append(fsls, fsl)
	}
	// Remove just in case it was left over from a previous run.
	for _, fn := range fns {
		ts.Remove(fn)
		mkFile(t, ts.FsLib, fn, HBUF, buf, SYNCFILESZ)
	}
	p1, err := perf.MakePerfMulti("TEST", "reader")
	assert.Nil(t, err)
	defer p1.Done()
	start := time.Now()
	for i := range fns {
		go func(i int) {
			n := measure(p1, "reader", func() sp.Tlength {
				r, err := fsls[i].OpenReader(fns[i])
				assert.Nil(t, err)
				n, err := test.Reader(t, r, buf, SYNCFILESZ)
				assert.Nil(t, err)
				r.Close()
				return n
			})
			done <- n
		}(i)
	}
	n := sp.Tlength(0)
	for _ = range fns {
		n += <-done
	}
	ms := time.Since(start).Milliseconds()
	db.DPrintf(db.ALWAYS, "Total tpt reader: %s took %vms (%s)", humanize.Bytes(uint64(n)), ms, test.TputStr(n, ms))
	for _, fn := range fns {
		err := ts.Remove(fn)
		assert.Nil(ts.T, err)
		mkFile(t, ts.FsLib, fn, HBUF, buf, FILESZ)
	}
	p2, err := perf.MakePerfMulti("TEST", "bufreader")
	assert.Nil(t, err)
	defer p2.Done()
	start = time.Now()
	for i := range fns {
		go func(i int) {
			n := measure(p2, "bufreader", func() sp.Tlength {
				r, err := fsls[i].OpenReader(fns[i])
				assert.Nil(t, err)
				br := bufio.NewReaderSize(r, sp.BUFSZ)
				n, err := test.Reader(t, br, buf, FILESZ)
				assert.Nil(t, err)
				r.Close()
				return n
			})
			done <- n
		}(i)
	}
	n = 0
	for _ = range fns {
		n += <-done
	}
	ms = time.Since(start).Milliseconds()
	db.DPrintf(db.ALWAYS, "Total tpt bufreader: %s took %vms (%s)", humanize.Bytes(uint64(n)), ms, test.TputStr(n, ms))
	p3, err := perf.MakePerfMulti("TEST", "abufreader")
	assert.Nil(t, err)
	defer p3.Done()
	start = time.Now()
	for i := range fns {
		go func(i int) {
			n := measure(p3, "readabuf", func() sp.Tlength {
				r, err := fsls[i].OpenAsyncReader(fns[i], 0)
				assert.Nil(t, err)
				n, err := test.Reader(t, r, buf, FILESZ)
				assert.Nil(t, err)
				r.Close()
				return n
			})
			done <- n
		}(i)
	}
	n = 0
	for _ = range fns {
		n += <-done
	}
	ms = time.Since(start).Milliseconds()
	db.DPrintf(db.ALWAYS, "Total tpt abufreader: %s took %vms (%s)", humanize.Bytes(uint64(n)), ms, test.TputStr(n, ms))
	ts.Shutdown()
}

func mkDir(t *testing.T, fsl *fslib.FsLib, dir string, n int) int {
	err := fsl.MkDir(dir, 0777)
	assert.Equal(t, nil, err)
	for i := 0; i < n; i++ {
		b := []byte("hello")
		_, err := fsl.PutFile(gopath.Join(dir, "f"+strconv.Itoa(i)), 0777, sp.OWRITE, b)
		assert.Nil(t, err)
	}
	return n
}

func TestDirCreatePerf(t *testing.T) {
	const N = 1000
	ts := test.MakeTstatePath(t, pathname)
	dir := gopath.Join(pathname, "d")
	measuredir("create dir", 1, func() int {
		n := mkDir(t, ts.FsLib, dir, N)
		return n
	})
	err := ts.RmDir(dir)
	assert.Nil(t, err)
	ts.Shutdown()
}

func lookuper(t *testing.T, nclerk int, n int, dir string, nfile int, nds []string) {
	const NITER = 100 // 10000
	ch := make(chan bool)
	for c := 0; c < nclerk; c++ {
		go func() {
			cn := strconv.Itoa(c)
			fsl, err := fslib.MakeFsLibAddr("fslibtest-"+cn, nds)
			assert.Nil(t, err)
			measuredir("lookup dir entry", NITER, func() int {
				for f := 0; f < nfile; f++ {
					_, err := fsl.Stat(gopath.Join(dir, "f"+strconv.Itoa(f)))
					assert.Nil(t, err)
				}
				return nfile
			})
			ch <- true
		}()
	}
	for c := 0; c < nclerk; c++ {
		<-ch
	}
}

func TestDirReadPerf(t *testing.T) {
	const N = 10000
	const NFILE = 10
	const NCLERK = 1
	ts := test.MakeTstatePath(t, pathname)
	dir := pathname + "d"
	n := mkDir(t, ts.FsLib, dir, NFILE)
	assert.Equal(t, NFILE, n)
	measuredir("read dir", 1, func() int {
		n := 0
		ts.ProcessDir(dir, func(st *sp.Stat) (bool, error) {
			n += 1
			return false, nil
		})
		return n
	})
	lookuper(t, 1, N, dir, NFILE, ts.NamedAddr())
	//lookuper(t, NCLERK, N, dir, NFILE)
	err := ts.RmDir(dir)
	assert.Nil(t, err)
	ts.Shutdown()
}

func TestRmDirPerf(t *testing.T) {
	const N = 5000
	ts := test.MakeTstatePath(t, pathname)
	dir := gopath.Join(pathname, "d")
	n := mkDir(t, ts.FsLib, dir, N)
	assert.Equal(t, N, n)
	measuredir("rm dir", 1, func() int {
		err := ts.RmDir(dir)
		assert.Nil(t, err)
		return N
	})
	ts.Shutdown()
}
