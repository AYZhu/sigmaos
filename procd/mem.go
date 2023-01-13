package procd

import (
	"io/ioutil"
	"strconv"
	"strings"

	db "sigmaos/debug"
	"sigmaos/proc"
)

func getMemTotal() proc.Tmem {
	b, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		db.DFatalf("Can't read /proc/meminfo: %v", err)
	}
	lines := strings.Split(string(b), "\n")
	for _, l := range lines {
		if strings.Contains(l, "MemTotal") {
			s := strings.Split(l, " ")
			kbStr := s[len(s)-2]
			kb, err := strconv.Atoi(kbStr)
			if err != nil {
				db.DFatalf("Couldn't convert MemTotal: %v", err)
			}
			return proc.Tmem(kb / 1000)
		}
	}
	db.DFatalf("Couldn't find total mem")
	return 0
}

func (pd *Procd) hasEnoughMemL(p *proc.Proc) bool {
	return pd.memAvail >= p.GetMem()
}

func (pd *Procd) allocMemL(p *proc.Proc) {
	pd.memAvail -= p.GetMem()
}

func (pd *Procd) freeMem(p *proc.Proc) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	pd.memAvail += p.GetMem()
}
