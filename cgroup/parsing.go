package cgroup

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"

	db "sigmaos/debug"
)

const (
	// The value comes from `C.sysconf(C._SC_CLK_TCK)`, and
	// on Linux it's a constant which is safe to be hard coded,
	// so we can avoid using cgo here. For details, see:
	// https://github.com/containerd/cgroups/pull/12
	clockTicksPerSecond   = 100
	microSecondsPerSecond = 1e6
)

type parseFn func(io.Reader) (uint64, error)

func parseUint64(r io.Reader) (uint64, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		db.DPrintf(db.CGROUP_ERR, "Error ReadAll: %v", err)
		return 0, err
	}
	n, err := strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64)
	if err != nil {
		db.DFatalf("Error strconv: %v", err)
	}
	return n, nil
}

func parseCgroupCpuStat(r io.Reader) (uint64, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		db.DPrintf(db.CGROUP_ERR, "Error ReadAll: %v", err)
		return 0, err
	}
	totalUsecsStr := strings.Fields(string(b))[1]
	totalUsecs, err := strconv.ParseUint(totalUsecsStr, 10, 64)
	if err != nil {
		db.DFatalf("Error strconv totalUsecs: %v", err)
	}
	return totalUsecs, nil
}

// Based on Docker's implementation:
// https://github.com/moby/moby/blob/master/daemon/stats/collector_unix.go
//
// getSystemCPUUsage returns the host system's cpu usage in
// microseconds. Uses /proc/stat defined by POSIX. Looks for the cpu
// statistics line and then sums up the first seven fields
// provided. See `man 5 proc` for details on specific field
// information.
func (cfs *cgroupFs) parseSysCpuStat(r io.Reader) (uint64, error) {
	if cfs.sysStatBr == nil {
		cfs.sysStatBr = bufio.NewReader(r)
	} else {
		cfs.sysStatBr.Reset(r)
	}
	defer cfs.sysStatBr.Reset(nil)

	for {
		line, err := cfs.sysStatBr.ReadString('\n')
		if err != nil {
			break
		}
		parts := strings.Fields(line)
		switch parts[0] {
		case "cpu":
			if len(parts) < 8 {
				db.DFatalf("invalid number of cpu fields %v", parts)
			}
			var totalClockTicks uint64
			for _, i := range parts[1:8] {
				v, err := strconv.ParseUint(i, 10, 64)
				if err != nil {
					db.DFatalf("Unable to convert value %s to int: %s", i, err)
				}
				totalClockTicks += v
			}
			return (totalClockTicks * microSecondsPerSecond) / clockTicksPerSecond, nil
		}
	}
	db.DFatalf("Error getSysCPUUsage")
	return 0, errors.New("Unexpected end of function parseSysCpuStat")
}
