package mr

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	// "github.com/klauspost/readahead"

	"ulambda/crash"
	db "ulambda/debug"
	"ulambda/delay"
	"ulambda/fslib"
	np "ulambda/ninep"
	"ulambda/proc"
	"ulambda/procclnt"
	"ulambda/rand"
	"ulambda/writer"
)

type Reducer struct {
	*fslib.FsLib
	*procclnt.ProcClnt
	reducef ReduceT
	input   string
	output  string
	tmp     string
	bwrt    *bufio.Writer
	wrt     *writer.Writer
}

func makeReducer(reducef ReduceT, args []string) (*Reducer, error) {
	if len(args) != 2 {
		return nil, errors.New("MakeReducer: too few arguments")
	}
	r := &Reducer{}
	r.input = args[0]
	r.output = args[1]
	r.tmp = r.output + rand.String(16)
	r.reducef = reducef
	r.FsLib = fslib.MakeFsLib("reducer-" + r.input)
	r.ProcClnt = procclnt.MakeProcClnt(r.FsLib)

	w, err := r.CreateWriter(r.tmp, 0777, np.OWRITE)
	if err != nil {
		return nil, err
	}
	r.wrt = w
	r.bwrt = bufio.NewWriterSize(w, BUFSZ)

	r.Started()

	crash.Crasher(r.FsLib)
	delay.SetDelayRPC(3)

	return r, nil
}

type result struct {
	kvs  []*KeyValue
	name string
	err  error
}

func (r *Reducer) readFile(ch chan result, file string) {
	// Make new fslib to parallelize request to a single fsux
	fsl := fslib.MakeFsLibAddr("r-"+file, fslib.Named())
	kvs := make([]*KeyValue, 0)
	d := r.input + "/" + file + "/"
	db.DPrintf("MR", "reduce %v\n", d)
	rdr, err := fsl.OpenReader(d)
	if err != nil {
		db.DPrintf("MR", "MakeReader %v err %v", d, err)
		ch <- result{nil, "", err}
		return
	}
	defer rdr.Close()
	defer fsl.Exit()

	brdr := bufio.NewReaderSize(rdr, BUFSZ)
	//ardr, err := readahead.NewReaderSize(rdr, 4, BUFSZ)
	//if err != nil {
	//	db.DFatalf("%v: readahead.NewReaderSize err %v", proc.GetName(), err)
	//}
	err = fslib.JsonReader(brdr, func() interface{} { return new(KeyValue) }, func(a interface{}) error {
		kv := a.(*KeyValue)
		db.DPrintf("REDUCE", "reduce %v: kv %v\n", file, kv)
		kvs = append(kvs, kv)
		return nil
	})
	if err != nil {
		ch <- result{nil, file, nil}
	} else {
		ch <- result{kvs, "", nil}
	}
}

func (r *Reducer) readFiles(input string) ([]*KeyValue, []string, error) {
	start := time.Now()
	kvs := []*KeyValue{}
	lostMaps := []string{}
	sts, err := r.GetDir(input)
	if err != nil {
		return nil, nil, err
	}
	ch := make(chan result)
	for _, st := range sts {
		go r.readFile(ch, st.Name)
	}
	for range sts {
		r := <-ch
		if r.err != nil {
			// another reducer already processed input
			// file; nothing to be done; bail out.
			return nil, nil, err
		}
		if r.name != "" {
			// If error is true, then either another
			// reducer already did the job (the input dir
			// is missing), the server holding the
			// mapper's output crashed, or is unreachable
			// (in which case we need to restart that
			// mapper).
			lostMaps = append(lostMaps, strings.TrimPrefix(r.name, "m-"))
		} else {
			kvs = append(kvs, r.kvs...)
		}
	}
	db.DPrintf("MR0", "Reduce Read %v\n", time.Since(start).Milliseconds())
	return kvs, lostMaps, nil
}

func (r *Reducer) emit(kv *KeyValue) error {
	b := fmt.Sprintf("%v %v\n", kv.Key, kv.Value)
	_, err := r.bwrt.Write([]byte(b))
	return err
}

func (r *Reducer) doReduce() *proc.Status {
	db.DPrintf(db.ALWAYS, "doReduce %v %v\n", r.input, r.output)
	kvs, lostMaps, err := r.readFiles(r.input)
	if err != nil {
		return proc.MakeStatusErr(fmt.Sprintf("%v: readFiles %v err %v\n", proc.GetName(), r.input, err), nil)
	}
	if len(lostMaps) > 0 {
		log.Printf("lost maps %v\n", lostMaps)
		return proc.MakeStatusErr(RESTART, lostMaps)
	}

	start := time.Now()
	sort.Sort(ByKey(kvs))
	db.DPrintf("MR0", "Reduce Sort %v\n", time.Since(start).Milliseconds())

	start = time.Now()
	i := 0
	for i < len(kvs) {
		j := i + 1
		for j < len(kvs) && kvs[j].Key == kvs[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, kvs[k].Value)
		}
		if err := r.reducef(kvs[i].Key, values, r.emit); err != nil {
			return proc.MakeStatusErr("reducef", err)
		}
		i = j
	}
	db.DPrintf("MR0", "Reduce reduce %v\n", time.Since(start).Milliseconds())

	start = time.Now()
	if err := r.bwrt.Flush(); err != nil {
		return proc.MakeStatusErr(fmt.Sprintf("%v: flush %v err %v\n", proc.GetName(), r.tmp, err), nil)
	}
	if err := r.wrt.Close(); err != nil {
		return proc.MakeStatusErr(fmt.Sprintf("%v: close %v err %v\n", proc.GetName(), r.tmp, err), nil)
	}
	err = r.Rename(r.tmp, r.output)
	if err != nil {
		return proc.MakeStatusErr(fmt.Sprintf("%v: rename %v -> %v err %v\n", proc.GetName(), r.tmp, r.output, err), nil)
	}
	db.DPrintf("MR0", "Reduce output %v\n", time.Since(start).Milliseconds())

	return proc.MakeStatus(proc.StatusOK)
}

func RunReducer(reducef ReduceT, args []string) {
	r, err := makeReducer(reducef, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: error %v", os.Args[0], err)
		os.Exit(1)
	}
	status := r.doReduce()
	r.Exited(status)
}
