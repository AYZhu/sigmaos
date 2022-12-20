package procd

import (
	"math"
	"runtime/debug"
	"time"

	db "sigmaos/debug"
	"sigmaos/fcall"
	"sigmaos/linuxsched"
	"sigmaos/proc"
	"sigmaos/resource"
	sp "sigmaos/sigmap"
)

type Tcorestatus uint8

const (
	CORE_AVAILABLE Tcorestatus = iota
	CORE_BLOCKED               // Not for use by this procd's procs.
)

func (st Tcorestatus) String() string {
	switch st {
	case CORE_AVAILABLE:
		return "CORE_AVAILABLE"
	case CORE_BLOCKED:
		return "CORE_BLOCKED"
	default:
		db.DFatalf("Unrecognized core status")
	}
	return ""
}

func (pd *Procd) initCores(grantedCoresIv string) {
	grantedCores := fcall.MkInterval(0, 0)
	grantedCores.Unmarshal(grantedCoresIv)
	// First, revoke access to all cores.
	allCoresIv := fcall.MkInterval(0, uint64(linuxsched.NCores))
	revokeMsg := resource.MakeResourceMsg(resource.Trequest, resource.Tcore, allCoresIv.Marshal(), int(linuxsched.NCores))
	pd.removeCores(revokeMsg)

	// Then, enable access to the granted core interval.
	grantMsg := resource.MakeResourceMsg(resource.Tgrant, resource.Tcore, grantedCores.Marshal(), int(grantedCores.Size()))
	pd.addCores(grantMsg)
}

func (pd *Procd) addCores(msg *resource.ResourceMsg) {
	pd.Lock()
	defer pd.Unlock()
	cores := parseCoreInterval(msg.Name)

	pd.adjustCoresOwned(pd.coresOwned, pd.coresOwned+proc.Tcore(msg.Amount), cores, CORE_AVAILABLE)
	for i := 0; i < msg.Amount; i++ {
		pd.wakeWorkerL()
	}
}

func (pd *Procd) removeCores(msg *resource.ResourceMsg) {
	pd.Lock()
	defer pd.Unlock()

	cores := parseCoreInterval(msg.Name)
	pd.adjustCoresOwned(pd.coresOwned, pd.coresOwned-proc.Tcore(msg.Amount), cores, CORE_BLOCKED)
}

func (pd *Procd) adjustCoresOwned(oldNCoresOwned, newNCoresOwned proc.Tcore, coresToMark []uint, newCoreStatus Tcorestatus) {
	// Mark cores according to their new status.
	pd.markCoresL(coresToMark, newCoreStatus)
	// Set the new procd core affinity.
	pd.setCoreAffinityL()
	// Rebalance procs given new cores.
	pd.rebalanceProcs(oldNCoresOwned, newNCoresOwned, coresToMark, newCoreStatus)
	pd.sanityCheckCoreCountsL()
}

// Rebalances procs across set of available cores. Procs are never evicted.
func (pd *Procd) rebalanceProcs(oldNCoresOwned, newNCoresOwned proc.Tcore, coresToMark []uint, newCoreStatus Tcorestatus) {
	// Free all procs' cores.
	for _, p := range pd.runningProcs {
		pd.freeCoresL(p)
	}
	// Sanity check
	if pd.coresAvail != oldNCoresOwned {
		pd.perf.Done()
		db.DFatalf("Mismatched num cores avail during rebalance: %v != %v", pd.coresAvail, oldNCoresOwned)
	}
	// Update the number of cores owned/available.
	pd.coresOwned = newNCoresOwned
	pd.coresAvail = newNCoresOwned
	// Calculate new core allocation for each proc, and track the
	// allocation. Rather than evict procs that don't fit, give them "0" cores.
	for _, p := range pd.runningProcs {
		newNCore := p.attr.Ncore
		// Make sure we don't overflow allocated cores.
		if newNCore > pd.coresAvail {
			newNCore = pd.coresAvail
		}
		// Resize the proc's core allocation.
		// Allocate cores to the proc.
		pd.allocCoresL(p, newNCore)
		// Set the CPU affinity for this proc to match procd.
		p.setCpuAffinityL()
	}
}

// Rate-limit how quickly we claim BE procs, since utilization statistics will
// take a while to update while claimed procs start. Return true if check
// passes and proc can be claimed.
//
// We claim a maximum of BE_PROC_OVERSUBSCRIPTION_RATE
// procs per underutilized core core per claim interval, where a claim interval
// is the length of ten CPU util samples.
func (pd *Procd) procClaimRateLimitCheck(util float64) bool {
	timeBetweenUtilSamples := time.Duration(1000/sp.Conf.Perf.CPU_UTIL_SAMPLE_HZ) * time.Millisecond
	// Check if we have moved onto the next interval (interval is currently 10 *
	// utilization sample rate).
	if time.Since(pd.procClaimTime) > 10*timeBetweenUtilSamples {
		pd.procClaimTime = time.Now()
		// We try to estimate the amount of "room" available for claiming new procs.
		pd.netProcsClaimed = proc.Tcore(math.Round(float64(pd.coresOwned) * util / 100.0))
		// If a proc is downloading, it's utilization won't have been measured yet.
		// Adding this to the number of procs claimed is perhaps a little too
		// conservative (we may double-count if the proc which is downloading was
		// also claimed in this epoch), but this should only happen the first time
		// a proc is downloaded, which should not be often.
		pd.netProcsClaimed += pd.procsDownloading
	}
	// If we have claimed < BE_PROC_OVERSUBSCRIPTION_RATE
	// procs per core during the last claim interval, the rate limit check
	// passes.
	maxOversub := proc.Tcore(sp.Conf.Procd.BE_PROC_OVERSUBSCRIPTION_RATE * float64(pd.coresOwned))
	if pd.netProcsClaimed < maxOversub {
		return true
	}
	db.DPrintf(db.PROCD, "Failed proc claim rate limit check: %v > %v", pd.netProcsClaimed, maxOversub)
	return false
}

func (pd *Procd) canClaimBEProc() bool {
	pd.Lock()
	defer pd.Unlock()

	_, _, ok := pd.canClaimBEProcL()
	return ok
}

func (pd *Procd) canClaimBEProcL() (float64, bool, bool) {
	// Determine whether or not we can run the proc based on
	// utilization and rate-limiting. If utilization is below a certain threshold,
	// take the proc.
	util, _ := pd.memfssrv.GetStats().GetUtil()
	rlc := pd.procClaimRateLimitCheck(util)
	if util < sp.Conf.Procd.BE_PROC_CLAIM_CPU_THRESHOLD && rlc {
		db.DPrintf(db.PROCD, "Have enough cores for BE proc: util %v rate-limit check %v", util, rlc)
		return util, rlc, true
	}
	db.DPrintf(db.PROCD, "Can't claim BE proc: util %v rate-limit check %v", util, rlc)
	return util, rlc, false
}

// Check if this procd has enough cores to run proc p. Caller holds lock.
func (pd *Procd) hasEnoughCores(p *proc.Proc) bool {
	// If this is an LC proc, check that we have enough cores.
	if p.Type == proc.T_LC {
		// If we have enough cores to run this job...
		if pd.coresAvail >= p.Ncore {
			return true
		}
		db.DPrintf(db.PROCD, "Don't have enough LC cores (%v) for %v", pd.coresAvail, p)
	} else {
		if util, rlc, ok := pd.canClaimBEProcL(); ok {
			return true
		} else {
			db.DPrintf(db.PROCD, "Not enough cores for BE proc: util %v rate-limit check %v proc %v", util, rlc, p)
		}
	}
	return false
}

// Allocate cores to a proc. Caller holds lock.
func (pd *Procd) allocCoresL(p *LinuxProc, n proc.Tcore) {
	if n > pd.coresAvail {
		debug.PrintStack()
		pd.perf.Done()
		db.DFatalf("Alloc too many cores %v %v", p, n)
	}
	p.coresAlloced = n
	pd.coresAvail -= n
	pd.netProcsClaimed++
	pd.sanityCheckCoreCountsL()
}

// Set the status of a set of cores. Caller holds lock.
func (pd *Procd) markCoresL(cores []uint, status Tcorestatus) {
	for _, i := range cores {
		// If we are double-setting a core's status, it's probably a bug.
		if pd.coreBitmap[i] == status {
			debug.PrintStack()
			pd.perf.Done()
			db.DFatalf("Error (noded:%v): Double-marked cores %v == %v", proc.GetNodedId(), pd.coreBitmap[i], status)
		}
		pd.coreBitmap[i] = status
	}
}

func (pd *Procd) freeCores(p *LinuxProc) {
	pd.Lock()
	defer pd.Unlock()

	pd.freeCoresL(p)
	if p.attr.Type != proc.T_LC {
		if pd.netProcsClaimed > 0 {
			pd.netProcsClaimed--
		}
	}
}

// Free a set of cores which was being used by a proc.
func (pd *Procd) freeCoresL(p *LinuxProc) {
	// If no cores were exclusively allocated to this proc, return immediately.
	if p.attr.Ncore == proc.C_DEF {
		return
	}

	pd.coresAvail += p.coresAlloced
	p.coresAlloced = 0
	pd.sanityCheckCoreCountsL()
}

func parseCoreInterval(ivStr string) []uint {
	iv := fcall.MkInterval(0, 0)
	iv.Unmarshal(ivStr)
	cores := make([]uint, iv.Size())
	for i := uint(0); i < uint(iv.Size()); i++ {
		cores[i] = uint(iv.Start) + i
	}
	return cores
}

// Run a sanity check for our core resource accounting. Caller holds lock.
func (pd *Procd) sanityCheckCoreCountsL() {
	if pd.coresOwned > proc.Tcore(linuxsched.NCores) {
		pd.perf.Done()
		db.DFatalf("Own more procd cores than there are cores on this machine: %v > %v", pd.coresOwned, linuxsched.NCores)
	}
	if pd.coresOwned < 0 {
		pd.perf.Done()
		db.DFatalf("Own too few cores: %v <= 0", pd.coresOwned)
	}
	if pd.coresAvail < 0 {
		pd.perf.Done()
		db.DFatalf("Too few cores available: %v < 0", pd.coresAvail)
	}
	if pd.coresAvail > pd.coresOwned {
		debug.PrintStack()
		pd.perf.Done()
		db.DFatalf("More cores available than cores owned: %v > %v", pd.coresAvail, pd.coresOwned)
	}
}
