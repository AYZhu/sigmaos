package proc

import (
	"path"
)

/*
 * Proc Directory structure:
 *
 * /
 * |- procd
 * |  |
 * |  |- x.x.x.x
 * |  |  |
 * |  |  |- pids
 * |  |     |
 * |  |     |- 1000 // Proc mounts this directory as procdir
 * |  |         |
 * |  |         |- evict-sem
 * |  |         |- children
 * |  |            |- 1001 // Child mounts this directory as procdir/parent
 * |  |               |- start-sem
 * |  |               |- status-pipe
 * |  |               |- procdir -> /procd/y.y.y.y/pids/1001 // Symlink to child's procdir.
 * |  |                  |- ...
 * |  |
 * |  |- y.y.y.y
 * |     |
 * |     |- pids
 * |        |
 * |        |- 1001
 * |            |
 * |            |- parent -> /procd/x.x.x.x/pids/1000/children/1001 // Mount of subdir of parent proc.
 * |            |- ...
 * |
 * |- kernel-pids // Only for kernel procs such as s3, ux, procd, ...
 *    |
 *    |- fsux-2000
 *       |
 *       |- kernel-proc // Only present if this is a kernel proc.
 *       |- ... // Same directory structure as regular procs
 */

const (
	// name for dir where procs live. May not refer to name/pids
	// because proc.PidDir may change it.  A proc refers to itself
	// using "pids/<pid>", where pid is the proc's PID.
	PIDS    = "pids" // TODO: make this explicitly kernel PIDs only
	PROCDIR = "procdir"

	// Files/directories in "pids/<pid>":
	START_SEM   = "start-sem"
	EVICT_SEM   = "evict-sem"
	RET_STATUS  = "status-pipe"
	CHILDREN    = "children"    // directory with children's pids and symlinks
	KERNEL_PROC = "kernel-proc" // Only present if this is a kernel proc
)

func GetChildProcDir(cpid string) string {
	return path.Join(GetProcDir(), CHILDREN, cpid, PROCDIR)
}

func GetChildStatusPath(cpid string) string {
	return path.Join(GetProcDir(), CHILDREN, cpid, RET_STATUS)
}
