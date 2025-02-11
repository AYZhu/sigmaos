package writer_test

import (
	"flag"
	gopath "path"
	"testing"

	"github.com/stretchr/testify/assert"

	sp "sigmaos/sigmap"
	"sigmaos/test"
)

var pathname string // e.g., --path "namedv1/"

func init() {
	flag.StringVar(&pathname, "path", sp.NAMED, "path for file system")
}

func TestWriter1(t *testing.T) {
	ts := test.MakeTstate(t)

	fn := gopath.Join(pathname, "f")
	d := []byte("abcdefg")
	wrt, err := ts.CreateWriter(fn, 0777, sp.OWRITE)
	assert.Nil(ts.T, err)

	for _, b := range d {
		v := make([]byte, 1)
		v[0] = b
		n, err := wrt.Write(v)
		assert.Equal(ts.T, nil, err)
		assert.Equal(ts.T, 1, n)
	}
	wrt.Close()

	d1, err := ts.GetFile(fn)
	assert.Equal(t, d, d1)

	err = ts.Remove(fn)
	assert.Nil(t, err, "Remove: %v", err)

	ts.Shutdown()
}
