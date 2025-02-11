package benchmarks_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	db "sigmaos/debug"
	"sigmaos/loadgen"
)

type HTTPClnt struct {
	url  string
	clnt *http.Client
}

func RunHTTPLoadGen(url string, dur time.Duration, maxrps int) {
	wc := makeHTTPClnt(url)
	lg := loadgen.MakeLoadGenerator(dur, maxrps, func(r *rand.Rand) {
		_, err := wc.get()
		if err != nil {
			db.DFatalf("Error HTTPLoadGen.Get: %v", err)
		}
	})
	db.DPrintf(db.TEST, "Calibrating loadgen")
	lg.Calibrate()
	db.DPrintf(db.TEST, "Running loadgen url %v dur %v maxrps %v", url, dur, maxrps)
	lg.Run()
	db.DPrintf(db.TEST, "Done generating load")
	lg.Stats()
}

func makeHTTPClnt(url string) *HTTPClnt {
	//	transport := &http.Transport{
	//		Dial: (&net.Dialer{
	//			Timeout: 2 * time.Minute,
	//		}).Dial,
	//	}
	clnt := &http.Client{
		Timeout:   2 * time.Minute,
		Transport: http.DefaultTransport,
	}
	// XXX This is sort of arbitrary, perhaps change or remove?.
	clnt.Transport.(*http.Transport).MaxIdleConnsPerHost = 10000
	return &HTTPClnt{url, clnt}
}

func (wc *HTTPClnt) get() (string, error) {
	resp, err := wc.clnt.Get(wc.url)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%v %s", resp.StatusCode, body)
	}
	return string(body), nil
}
