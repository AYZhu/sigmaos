package proc

import (
	"path"
)

/*
 * Proc Directory structure:
 *
 * /
 * |- schedd
 * |  |
 * |  |- kernel-1
 * |  |  |
 * |  |  |- pids
 * |  |     |
 * |  |     |- 1000 // Proc mounts this directory as procdir
 * |  |         |
 * |  |         |- evict-sem
 * |  |         |- exit-sem
 * |  |         |- children
 * |  |            |- 1001 // Child mounts this directory as procdir/parent
 * |  |               |- start-sem
 * |  |               |- exit-status
 * |  |               |- shared -> link/to/parent/shared/state // Symlink to shared state of parent's choosing, if desired.
 * |  |               |- procdir -> /schedd/kernel-2/pids/1001 // Symlink to child's procdir.
 * |  |                  |- ...
 * |  |
 * |  |- kernel-2
 * |     |
 * |     |- pids
 * |        |
 * |        |- 1001
 * |            |
 * |            |- parentdir -> /schedd/kernel-1/pids/1000/children/1001 // Mount of subdir of parent proc.
 * |            |- ...
 * |
 * |- kpids // Only for kernel procs such as s3, ux, schedd, ...
 *    |
 *    |- schedd-2000
 *       |
 *       |- kernel-proc // Only present if this is a kernel proc.
 *       |- ... // Same directory structure as regular procs
 */

const (
	// name for dir where procs live. May not refer to name/pids
	// because proc.PidDir may change it.  A proc refers to itself
	// using "pids/<pid>", where pid is the proc's PID.
	PROCDIR       = "procdir"
	PARENTDIR     = "parentdir"
	PROCFILE_LINK = "procfile-link"

	// Files/directories in "pids/<pid>":
	SHARED      = "shared"
	START_SEM   = "start-sem"
	EXIT_SEM    = "exit-sem"
	EVICT_SEM   = "evict-sem"
	EXIT_STATUS = "exit-status"
	CHILDREN    = "children" // directory with children's pids and symlinks
)

func GetChildProcDir(procdir string, cpid Tpid) string {
	return path.Join(procdir, CHILDREN, cpid.String(), PROCDIR)
}
