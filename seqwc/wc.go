package seqwc

import (
	"bufio"
	"fmt"
	// "encoding/json"
	"io"
	"log"
	"strings"
	"time"

	"github.com/dustin/go-humanize"

	db "sigmaos/debug"
	"sigmaos/fslib"
	"sigmaos/mr"
	"sigmaos/perf"
	sp "sigmaos/sigmap"
	"sigmaos/test"
)

type Tdata map[string]uint64

func Wcline(n int, line string, data Tdata, sbc *mr.ScanByteCounter) int {
	scanner := bufio.NewScanner(strings.NewReader(line))
	scanner.Split(sbc.ScanWords)
	cnt := 0
	for scanner.Scan() {
		w := scanner.Text()
		if _, ok := data[w]; !ok {
			data[w] = uint64(0)
		}
		// kv := &mr.KeyValue{scanner.Text(), "1"}
		// _, err := json.Marshal(kv)
		// if err != nil {
		// 	db.DFatalf("json")
		// }
		data[w] += 1
		cnt++
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("scanner err %v\n", err)
	}
	return cnt
}

func wcFile(rdr io.Reader, data Tdata, sbc *mr.ScanByteCounter) int {
	scanner := bufio.NewScanner(rdr)
	buf := make([]byte, 0, 8*sp.MBYTE)
	scanner.Buffer(buf, cap(buf))
	n := 1
	cnt := 0
	for scanner.Scan() {
		l := scanner.Text()
		cnt += Wcline(n, l, data, sbc)
		n += 1
	}
	return cnt
}

func WcData(fsl *fslib.FsLib, dir string, data Tdata, sbc *mr.ScanByteCounter) (int, sp.Tlength, error) {
	sts, err := fsl.GetDir(dir)
	if err != nil {
		return 0, 0, err
	}
	n := 0
	nbytes := sp.Tlength(0)
	for _, st := range sts {
		nbytes += st.Tlength()
		rdr, err := fsl.OpenAsyncReader(dir+"/"+st.Name, 0)
		if err != nil {
			return 0, 0, err
		}
		m := wcFile(rdr, data, sbc)
		// log.Printf("%v: %d\n", st.Name, m)
		n += m
	}
	return n, nbytes, nil
}

func Wc(fsl *fslib.FsLib, dir string, out string) (int, error) {
	p, err := perf.MakePerf(fsl.SigmaConfig(), perf.SEQWC)
	if err != nil {
		return 0, err
	}
	sbc := mr.MakeScanByteCounter(p)
	data := make(Tdata)
	start := time.Now()
	n, nbytes, err := WcData(fsl, dir, data, sbc)
	wrt, err := fsl.CreateAsyncWriter(out, 0777, sp.OWRITE|sp.OTRUNC)
	if err != nil {
		return 0, err
	}
	defer wrt.Close()
	for k, v := range data {
		b := fmt.Sprintf("%s\t%d\n", k, v)
		_, err := wrt.Write([]byte(b))
		if err != nil {
			return 0, err
		}
	}

	ms := time.Since(start).Milliseconds()
	db.DPrintf(db.ALWAYS, "Wc %s took %vms (%s)", humanize.Bytes(uint64(nbytes)), ms, test.TputStr(nbytes, ms))
	return n, nil
}
