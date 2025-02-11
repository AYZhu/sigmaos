package fss3

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	db "sigmaos/debug"
	"sigmaos/fs"
	"sigmaos/path"
	"sigmaos/serr"
	sp "sigmaos/sigmap"
	"sigmaos/sorteddir"
	"sigmaos/spcodec"
)

const DOT = "_._"

func toDot(pn string) string {
	path := path.Split(pn)
	if len(path) > 0 && path.Base() == "." {
		path[len(path)-1] = DOT
	}
	return path.String()
}

func fromDot(pn string) string {
	return strings.Replace(pn, DOT, ".", -1)
}

type Dir struct {
	*Obj
	sync.Mutex
	dents *sorteddir.SortedDir
	sts   []*sp.Stat
}

func (d *Dir) String() string {
	s := d.Obj.String()
	return s + fmt.Sprintf(" dents %v", d.dents)
}

func makeDir(bucket string, key path.Path, perm sp.Tperm) *Dir {
	o := makeObj(bucket, key, perm)
	dir := &Dir{}
	dir.Obj = o
	dir.dents = sorteddir.MkSortedDir()
	return dir
}

func (d *Dir) readRoot(ctx fs.CtxI) *serr.Err {
	db.DPrintf(db.S3, "readRoot %v\n", d)
	input := &s3.ListBucketsInput{}
	result, err := fss3.client.ListBuckets(context.TODO(), input)
	if err != nil {
		return serr.MkErr(serr.TErrError, err)
	} else {
		for _, b := range result.Buckets {
			d.dents.Insert(strings.TrimRight(*b.Name, "/"), sp.DMDIR)
		}
	}
	db.DPrintf(db.S3, "readRoot: dir %v\n", d)
	return d.statDir(ctx)
}

// lookup bucket in root directory
func (d *Dir) lookupBucket(ctx fs.CtxI, b string) *serr.Err {
	if err := d.fill(ctx); err != nil {
		return err
	}
	_, ok := d.dents.Lookup(b)
	if !ok {
		return serr.MkErr(serr.TErrNotfound, b)
	}
	return nil
}

func (d *Dir) s3ReadDir(ctx fs.CtxI, fss3 *Fss3) *serr.Err {
	maxKeys := 0
	if d.bucket == "" { // root dir?
		return d.readRoot(ctx)
	}
	key := d.key.String()
	if len(d.key) > 0 && d.key.Base() == "." {
		key = d.key[:len(d.key)-1].String()
	}
	if key != "" {
		key = key + "/"
	}
	params := &s3.ListObjectsV2Input{
		Bucket:    &d.bucket,
		Prefix:    aws.String(key),
		Delimiter: aws.String("/"),
	}
	db.DPrintf(db.S3, "s3ReadDir %v params %v\n", d, params)
	p := s3.NewListObjectsV2Paginator(fss3.client, params,
		func(o *s3.ListObjectsV2PaginatorOptions) {
			if v := int32(maxKeys); v != 0 {
				o.Limit = v
			}
		})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return serr.MkErr(serr.TErrInval, err)
		}
		for _, obj := range page.Contents {
			db.DPrintf(db.S3, "key %v\n", *obj.Key)
			n := strings.TrimPrefix(*obj.Key, key)
			if n == DOT {
				d.dents.Insert(".", sp.DMDIR)
			} else {
				d.dents.Insert(n, sp.Tperm(0777))
			}
		}
		for _, obj := range page.CommonPrefixes {
			db.DPrintf(db.S3, "prefix %v\n", *obj.Prefix)
			n := strings.TrimPrefix(*obj.Prefix, key)
			d.dents.Insert(strings.TrimRight(n, "/"), sp.DMDIR)
		}
	}
	d.sz = sp.Tlength(d.dents.Len()) // makeup size
	db.DPrintf(db.S3, "s3ReadDir: dir %v key %v\n", d, key)
	return nil
}

func (d *Dir) fill(ctx fs.CtxI) *serr.Err {
	if d.sz > 0 { // already filled?
		return nil
	}
	if err := d.s3ReadDir(ctx, fss3); err != nil {
		return err
	}
	return nil
}

func (d *Dir) dirents() []*Obj {
	d.Lock()
	defer d.Unlock()
	dents := make([]*Obj, 0, d.dents.Len())
	d.dents.Iter(func(n string, e interface{}) bool {
		if n != "." {
			dents = append(dents, makeObj(d.bucket, d.key.Copy().Append(n), e.(sp.Tperm)))
		}
		return true
	})
	return dents
}

func (d *Dir) Stat(ctx fs.CtxI) (*sp.Stat, *serr.Err) {
	d.Lock()
	defer d.Unlock()
	db.DPrintf(db.S3, "Stat dir %v\n", d)
	if err := d.fill(ctx); err != nil {
		return nil, err
	}
	st := d.stat()
	st.Length = uint64(d.sz)
	return st, nil
}

func mkObjs(base *Obj) []fs.FsObj {
	os := make([]fs.FsObj, 0, len(base.key))
	for i, _ := range base.key {
		if i+1 >= len(base.key) {
			break
		}
		os = append(os, makeFsObj(base.bucket, sp.DMDIR, base.key[0:i+1]))
	}
	return os
}

func (d *Dir) lookupPath(ctx fs.CtxI, p path.Path) ([]fs.FsObj, fs.FsObj, path.Path, *serr.Err) {
	db.DPrintf(db.S3, "%v: lookupPath d %v p %v\n", ctx, d, p)
	// maybe p is f a file
	o := makeObj(d.bucket, d.key.Copy().AppendPath(p), sp.Tperm(0777))
	if err := o.readHead(fss3); err == nil {
		// name is a file; done
		db.DPrintf(db.S3, "Lookup return %q o %v\n", p, o)
		os := append(mkObjs(o), o)
		return os, o, nil, nil
	}

	// maybe p is a dir
	d1 := makeDir(d.bucket, d.key.Copy().AppendPath(p), sp.DMDIR|sp.Tperm(0777))
	if err := d1.fill(ctx); err != nil {
		db.DPrintf(db.S3, "Lookup %q err %v\n", p, err)
		return nil, nil, p, err
	}
	if d1.dents.Len() == 0 {
		// not a directory either
		db.DPrintf(db.S3, "Lookup %q not found\n", p)
		return nil, nil, p, serr.MkErr(serr.TErrNotfound, p)
	}
	db.DPrintf(db.S3, "Lookup return %q %v\n", p, d1)
	return append(mkObjs(d1.Obj), d1), d1, nil, nil
}

func (d *Dir) LookupPath(ctx fs.CtxI, p path.Path) ([]fs.FsObj, fs.FsObj, path.Path, *serr.Err) {
	db.DPrintf(db.S3, "%v: LookupPath d %v %v\n", ctx, d, p)
	// if d is the root directory, then the first pathname component
	// is a bucket name, resolve it first.
	if d.bucket == "" {
		if err := d.lookupBucket(ctx, p[0]); err != nil {
			return nil, nil, p, err
		}
		d1 := makeDir(p[0], path.Path{}, sp.DMDIR|sp.Tperm(0777))
		return []fs.FsObj{d1.Obj}, d1, p[1:], nil
	}
	return d.lookupPath(ctx, p)
}

func (d *Dir) statDir(ctx fs.CtxI) *serr.Err {
	d.sts = make([]*sp.Stat, 0, d.dents.Len())
	for _, o := range d.dirents() {
		var st *sp.Stat
		var err *serr.Err
		if o.perm.IsDir() {
			st = o.stat()
		} else {
			st, err = o.Stat(ctx)
		}
		if err != nil {
			// another proc may have removed the file
			d.dents.Delete(o.key.Base())
			continue
		}
		d.sts = append(d.sts, st)
	}
	sz, err := spcodec.MarshalSizeDir(d.sts)
	if err != nil {
		return err
	}
	d.sz = sz
	return nil
}

func (d *Dir) Open(ctx fs.CtxI, m sp.Tmode) (fs.FsObj, *serr.Err) {
	db.DPrintf(db.S3, "open dir %v (%T) %v\n", d, d, m)
	if err := d.fill(ctx); err != nil {
		return nil, err
	}
	if err := d.statDir(ctx); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *Dir) ReadDir(ctx fs.CtxI, cursor int, cnt sp.Tsize) ([]*sp.Stat, *serr.Err) {
	db.DPrintf(db.S3, "ReadDir %v %d cursor %d cnt %v\n", d, len(d.sts), cursor, cnt)

	if err := d.fill(ctx); err != nil {
		return nil, err
	}
	if cursor > len(d.sts) {
		return nil, nil
	} else {
		return d.sts[cursor:], nil
	}
}

// Create a fake file in dir to materialize dir
func (d *Dir) CreateDir(ctx fs.CtxI, name string, perm sp.Tperm) (fs.FsObj, *serr.Err) {
	key := d.key.Copy().Append(name).Append(DOT).String()
	db.DPrintf(db.S3, "CreateDir: %v\n", key)
	input := &s3.PutObjectInput{
		Bucket: &d.bucket,
		Key:    &key,
	}
	_, err := fss3.client.PutObject(context.TODO(), input)
	if err != nil {
		return nil, serr.MkErrError(err)
	}
	o := makeFsObj(d.bucket, perm, d.key.Copy().Append(name))
	return o, nil
}

func (d *Dir) Create(ctx fs.CtxI, name string, perm sp.Tperm, m sp.Tmode, lid sp.TleaseId, f sp.Tfence) (fs.FsObj, *serr.Err) {
	db.DPrintf(db.S3, "Create %v name: %v\n", d, name)
	if d.bucket == "" {
		return nil, serr.MkErr(serr.TErrNocreate, d.bucket)
	}
	o := makeObj(d.bucket, d.key.Copy().Append(name), perm)
	_, err := o.Stat(ctx)
	if err == nil {
		return nil, serr.MkErr(serr.TErrExists, name)
	}
	if perm.IsDir() {
		obj, err := d.CreateDir(ctx, name, perm)
		if err != nil {
			return nil, err
		}
		d.dents.Insert(name, perm)
		return obj, nil
	}
	d.dents.Insert(name, perm)
	if err := o.s3Create(); err != nil {
		return nil, err
	}
	if perm.IsFile() && m == sp.OWRITE {
		o.setupWriter()
	}
	return o, nil
}

func (d *Dir) Renameat(ctx fs.CtxI, from string, od fs.Dir, to string, f sp.Tfence) *serr.Err {
	return serr.MkErr(serr.TErrNotSupported, "Renameat")
}

func (d *Dir) Remove(ctx fs.CtxI, name string, f sp.Tfence) *serr.Err {
	if d.bucket == "" {
		return serr.MkErr(serr.TErrNoremove, d.bucket)
	}
	key := d.key.Copy().Append(name)
	if err := d.fill(ctx); err != nil {
		return err
	}
	db.DPrintf(db.S3, "Delete %v key %v name %v\n", d, key, name)
	e, ok := d.dents.Lookup(name)
	if !ok {
		db.DPrintf(db.S3, "Delete %v err %v\n", key, name)
		return serr.MkErr(serr.TErrNotfound, name)
	}
	perm := e.(sp.Tperm)
	if perm.IsDir() {
		d1 := makeDir(d.bucket, d.key.Copy().Append(name), perm)
		if err := d1.s3ReadDir(ctx, fss3); err != nil {
			return err
		}
		if d1.dents.Len() > 1 {
			return serr.MkErr(serr.TErrNotEmpty, name)
		}
		key = key.Append(DOT)
	}
	k := key.String()
	input := &s3.DeleteObjectInput{
		Bucket: &d.bucket,
		Key:    &k,
	}
	if _, err := fss3.client.DeleteObject(context.TODO(), input); err != nil {
		db.DPrintf(db.S3, "DeleteObject %v err %v\n", k, err)
		return serr.MkErrError(err)
	}
	d.dents.Delete(name)
	return nil
}

func (d *Dir) Rename(ctx fs.CtxI, from, to string, f sp.Tfence) *serr.Err {
	return serr.MkErr(serr.TErrNotSupported, "Rename")
}

// ===== The following functions are needed to make an s3 dir of type fs.Inode

func (d *Dir) SetMtime(mtime int64) {
	db.DFatalf("Unimplemented")
}

func (d *Dir) Mtime() int64 {
	db.DFatalf("Unimplemented")
	return 0
}

func (d *Dir) SetParent(di fs.Dir) {
	db.DFatalf("Unimplemented")
}

func (d *Dir) Unlink() {
	db.DFatalf("Unimplemented")
}

func (d *Dir) VersionInc() {
	db.DFatalf("Unimplemented")
}
