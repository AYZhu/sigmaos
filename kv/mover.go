package kv

import (
	"fmt"
	"strconv"
	"sync"

	"sigmaos/crash"
	db "sigmaos/debug"
	"sigmaos/fslib"
	"sigmaos/proc"
	"sigmaos/serr"
	"sigmaos/sessp"
	"sigmaos/sigmaclnt"
	sp "sigmaos/sigmap"
)

//
// Move shards between servers
//

type Mover struct {
	mu sync.Mutex
	*sigmaclnt.SigmaClnt
	job   string
	fence *sessp.Tfence
	shard uint32
	cc    *CacheClnt
}

func JoinEpoch(fsl *fslib.FsLib, job string, fence *sessp.Tfence) error {
	config := Config{}
	if err := fsl.GetFileJson(KVConfig(job), &config); err != nil {
		return fmt.Errorf("GetFileJson %v err %v", KVConfig(job), err)
	}
	if config.Fence.Epoch != fence.Epoch {
		return fmt.Errorf("Newer config %v", config.Fence)
	}
	return nil
}

func MakeMover(job, epochstr, shard, src, dst string) (*Mover, error) {
	sc, err := sigmaclnt.MkSigmaClnt(sp.Tuname("mover-" + proc.GetPid().String()))
	if err != nil {
		return nil, err
	}
	fence, err := sessp.NewFenceJson([]byte(epochstr))
	if err != nil {
		return nil, err
	}
	mv := &Mover{fence: fence, SigmaClnt: sc, job: job, cc: NewCacheClnt(sc.FsLib, NSHARD)}
	if sh, err := strconv.ParseUint(shard, 10, 32); err != nil {
		return nil, err
	} else {
		mv.shard = uint32(sh)
	}
	if err := mv.Started(); err != nil {
		db.DFatalf("%v: couldn't start %v\n", proc.GetName(), err)
	}
	crash.Crasher(mv.FsLib)

	// if err := JoinEpoch(mv.FsLib, mv.job, mv.fence); err != nil {
	// 	mv.ClntExit(proc.MakeStatusErr(err.Error(), nil))
	// 	return nil, err
	// }
	return mv, nil
}

// Copy shard from src to dst
func (mv *Mover) moveShard(s, d string) error {
	if err := mv.cc.FreezeShard(s, mv.shard, mv.fence); err != nil {
		db.DPrintf(db.KVMV, "FreezeShard %v err %v\n", s, err)
		// did previous mover finish the job?
		if serr.IsErrCode(err, serr.TErrNotfound) {
			return nil
		}
		return err
	}

	// A crashed mover may have created the shard and partially filled
	// it; remove it and start over.
	if err := mv.cc.DeleteShard(d, mv.shard, mv.fence); err != nil {
		db.DPrintf(db.KVMV, "DeleteShard %v err %v\n", mv.shard, err)
		if !serr.IsErrCode(err, serr.TErrNotfound) {
			return err
		}
	}

	if err := mv.cc.CreateShard(d, mv.shard, mv.fence); err != nil {
		db.DPrintf(db.KVMV, "CreateShard %v err %v\n", mv.shard, err)
		return err
	}
	vals, err := mv.cc.DumpShard(s, mv.shard)
	if err != nil {
		db.DPrintf(db.KVMV, "DumpShard %v err %v\n", mv.shard, err)
		return err
	}
	if err := mv.cc.FillShard(d, mv.shard, vals, mv.fence); err != nil {
		db.DPrintf(db.KVMV, "FillShard %v err %v\n", mv.shard, err)
		return err
	}

	// Mark that move is done by deleting s
	if err := mv.cc.DeleteShard(s, mv.shard, mv.fence); err != nil {
		db.DPrintf(db.KVMV, "DeleteShard %v err %v\n", mv.shard, err)
		return err
	}
	return nil
}

func (mv *Mover) Move(src, dst string) {
	db.DPrintf(db.KVMV, "conf %v: mv %d from %v to %v\n", mv.fence, mv.shard, src, dst)
	err := mv.moveShard(src, dst)
	if err != nil {
		db.DPrintf(db.KVMV_ERR, "conf %v: mv %d from %v to %v err %v\n", mv.fence, mv.shard, src, dst, err)
	}
	db.DPrintf(db.KVMV, "conf %v: mv %d  done from %v to %v err %v\n", mv.fence, mv.shard, src, dst, err)
	if err != nil {
		mv.ClntExit(proc.MakeStatusErr(err.Error(), nil))
	} else {
		mv.ClntExitOK()
	}
}
