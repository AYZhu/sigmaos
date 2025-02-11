package imgresized

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"sigmaos/crash"
	db "sigmaos/debug"
	"sigmaos/fslib"
	"sigmaos/leaderclnt"
	"sigmaos/proc"
	rd "sigmaos/rand"
	"sigmaos/serr"
	"sigmaos/sigmaclnt"
	sp "sigmaos/sigmap"
)

const (
	IMG    = "name/img"
	STOP   = "_STOP"
	NCOORD = 1
)

type ImgSrv struct {
	*sigmaclnt.SigmaClnt
	job        string
	done       string
	wip        string
	todo       string
	workerMcpu proc.Tmcpu
	isDone     bool
	crash      int
	leaderclnt *leaderclnt.LeaderClnt
}

func MkDirs(fsl *fslib.FsLib, job string) error {
	fsl.RmDir(IMG)
	if err := fsl.MkDir(IMG, 0777); err != nil {
		return err
	}
	if err := fsl.MkDir(path.Join(IMG, job), 0777); err != nil {
		return err
	}
	if err := fsl.MkDir(path.Join(IMG, job, "done"), 0777); err != nil {
		return err
	}
	if err := fsl.MkDir(path.Join(IMG, job, "todo"), 0777); err != nil {
		return err
	}
	if err := fsl.MkDir(path.Join(IMG, job, "wip"), 0777); err != nil {
		return err
	}
	return nil
}

func SubmitTask(fsl *fslib.FsLib, job string, fn string) error {
	return SubmitTaskMulti(fsl, job, []string{fn})
}

func SubmitTaskMulti(fsl *fslib.FsLib, job string, fns []string) error {
	t := path.Join(sp.IMG, job, "todo", rd.String(4))
	_, err := fsl.PutFile(t, 0777, sp.OREAD, []byte(strings.Join(fns, ",")))
	return err
}

func NTaskDone(fsl *fslib.FsLib, job string) (int, error) {
	sts, err := fsl.GetDir(path.Join(sp.IMG, job, "done"))
	if err != nil {
		return -1, err
	}
	return len(sts), nil
}

func Cleanup(fsl *fslib.FsLib, dir string) error {
	_, err := fsl.ProcessDir(dir, func(st *sp.Stat) (bool, error) {
		if strings.Contains(st.Name, "thumb") {
			err := fsl.Remove(path.Join(dir, st.Name))
			if err != nil {
				return true, err
			}
			return false, nil
		}
		return false, nil
	})
	return err
}

func MakeImgd(args []string) (*ImgSrv, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("MakeImgSrv: wrong number of arguments: %v", args)
	}
	imgd := &ImgSrv{}
	imgd.job = args[0]
	sc, err := sigmaclnt.MkSigmaClnt(sp.Tuname("imgd-" + proc.GetPid().String()))
	if err != nil {
		return nil, err
	}
	db.DPrintf(db.IMGD, "Made fslib job %v, addr %v", imgd.job, sc.NamedAddr())
	imgd.SigmaClnt = sc
	imgd.job = args[0]
	crashing, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("MakeImgSrv: error parse crash %v", err)
	}
	imgd.crash = crashing
	imgd.done = path.Join(IMG, imgd.job, "done")
	imgd.todo = path.Join(IMG, imgd.job, "todo")
	imgd.wip = path.Join(IMG, imgd.job, "wip")
	mcpu, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, fmt.Errorf("MakeImgSrv: Error parse MCPU %v", err)
	}
	imgd.workerMcpu = proc.Tmcpu(mcpu)

	imgd.Started()

	imgd.leaderclnt, err = leaderclnt.MakeLeaderClnt(imgd.FsLib, path.Join(IMG, imgd.job, "imgd-leader"), 0777)
	if err != nil {
		return nil, fmt.Errorf("MakeLeaderclnt err %v", err)
	}

	crash.Crasher(imgd.FsLib)

	go func() {
		imgd.WaitEvict(proc.GetPid())
		imgd.ClntExitOK()
		os.Exit(0)
	}()

	return imgd, nil
}

func (imgd *ImgSrv) claimEntry(name string) (string, error) {
	if err := imgd.Rename(imgd.todo+"/"+name, imgd.wip+"/"+name); err != nil {
		if serr.IsErrCode(err, serr.TErrUnreachable) { // partitioned?
			return "", err
		}
		// another thread claimed the task before us
		return "", nil
	}
	return name, nil
}

type task struct {
	name string
	fn   string
}

type Tresult struct {
	t   string
	ok  bool
	ms  int64
	msg string
}

func (imgd *ImgSrv) waitForTask(start time.Time, ch chan Tresult, p *proc.Proc, t task) {
	imgd.WaitStart(p.GetPid())
	db.DPrintf(db.ALWAYS, "Start Latency %v", time.Since(start))
	status, err := imgd.WaitExit(p.GetPid())
	ms := time.Since(start).Milliseconds()
	if err == nil && status.IsStatusOK() {
		// mark task as done
		if err := imgd.Rename(imgd.wip+"/"+t.name, imgd.done+"/"+t.name); err != nil {
			db.DFatalf("rename task %v done err %v\n", t, err)
		}
		ch <- Tresult{t.name, true, ms, status.Msg()}
	} else { // task failed; make it runnable again
		db.DPrintf(db.IMGD, "task %v failed %v err %v\n", t, status, err)
		if err := imgd.Rename(imgd.wip+"/"+t.name, imgd.todo+"/"+t.name); err != nil {
			db.DFatalf("rename task %v todo err %v\n", t, err)
		}
		ch <- Tresult{t.name, false, ms, ""}
	}
}

func ThumbName(fn string) string {
	ext := path.Ext(fn)
	fn1 := strings.TrimSuffix(fn, ext) + "-" + rd.String(4) + "-thumb" + path.Ext(fn)
	return fn1
}

func (imgd *ImgSrv) runTasks(ch chan Tresult, tasks []task) {
	procs := make([]*proc.Proc, len(tasks))
	for i, t := range tasks {
		procs[i] = proc.MakeProc("imgresize", []string{t.fn, ThumbName(t.fn)})
		if imgd.crash > 0 {
			procs[i].AppendEnv("SIGMACRASH", strconv.Itoa(imgd.crash))
		}
		procs[i].SetMcpu(imgd.workerMcpu)
		db.DPrintf(db.IMGD, "prep to burst-spawn task %v %v\n", procs[i].GetPid(), procs[i].Args)
	}
	start := time.Now()
	// Burst-spawn procs.
	failed, errs := imgd.SpawnBurst(procs, 1)
	if len(failed) > 0 {
		db.DFatalf("Couldn't burst-spawn some tasks %v, errs: %v", failed, errs)
	}
	for i := range procs {
		go imgd.waitForTask(start, ch, procs[i], tasks[i])
	}
}

func (imgd *ImgSrv) work(sts []*sp.Stat) bool {
	if imgd.isDone {
		return false
	}
	tasks := []task{}
	ch := make(chan Tresult)
	for _, st := range sts {
		t, err := imgd.claimEntry(st.Name)
		if err != nil || t == "" {
			continue
		}
		s3fn, err := imgd.GetFile(path.Join(imgd.wip, t))
		if err != nil {
			continue
		}
		if string(s3fn) == STOP {
			return false
		}
		tasks = append(tasks, task{t, string(s3fn)})
	}
	go imgd.runTasks(ch, tasks)
	for i := len(tasks); i > 0; i-- {
		res := <-ch
		if res.ok {
			db.DPrintf(db.IMGD, "%v ok %v ms %d msg %v\n", res.t, res.ok, res.ms, res.msg)
			//if err := c.AppendFileJson(MRstats(c.job), res.res); err != nil {
			//	db.DFatalf("Appendfile %v err %v\n", MRstats(c.job), err)
			//}
		}
	}
	return true
}

// Consider all tasks in progress as failed (too aggressive, but
// correct), and make them runnable
func (imgd *ImgSrv) recover() {
	if _, err := imgd.MoveFiles(imgd.wip, imgd.todo); err != nil {
		db.DFatalf("MoveFiles %v err %v\n", imgd.wip, err)
	}
}

func (imgd *ImgSrv) Work() {

	db.DPrintf(db.IMGD, "Try acquire leadership coord %v job %v", proc.GetPid(), imgd.job)

	// Try to become the leading coordinator.
	if err := imgd.leaderclnt.LeadAndFence(nil, []string{path.Join(IMG, imgd.job)}); err != nil {
		db.DFatalf("LeadAndFence err %v", err)
	}

	db.DPrintf(db.ALWAYS, "leader %s\n", imgd.job)

	imgd.recover()

	work := true
	for work {
		sts, err := imgd.ReadDirWatch(imgd.todo, func(sts []*sp.Stat) bool {
			return len(sts) == 0
		})
		if err != nil {
			db.DFatalf("ReadDirWatch %v err %v\n", imgd.todo, err)
		}
		work = imgd.work(sts)
	}

	db.DPrintf(db.ALWAYS, "imgresized exit\n")

	imgd.ClntExitOK()
}
