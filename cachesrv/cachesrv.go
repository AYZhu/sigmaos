package cachesrv

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	cacheproto "sigmaos/cache/proto"

	"sigmaos/cache"
	db "sigmaos/debug"
	"sigmaos/fs"
	"sigmaos/perf"
	"sigmaos/proc"
	"sigmaos/repl"
	"sigmaos/replraft"
	"sigmaos/serr"
	"sigmaos/sessdevsrv"
	sp "sigmaos/sigmap"
	"sigmaos/sigmasrv"
	"sigmaos/tracing"
)

const (
	DUMP   = "dump"
	NSHARD = 1009 // for cached
)

type Tstatus string

const (
	READY  Tstatus = "Ready"
	FROZEN Tstatus = "Frozen"
)

type shardInfo struct {
	status Tstatus
	s      *shard
}

type shardMap map[cache.Tshard]*shardInfo

type CacheSrv struct {
	mu        sync.Mutex
	shards    shardMap
	shrd      string
	raftcfg   *replraft.RaftConfig
	replSrv   repl.Server
	tracer    *tracing.Tracer
	lastFence *sp.Tfence
	perf      *perf.Perf
}

func RunCacheSrv(args []string, nshard int) error {
	pn := ""
	if len(args) > 3 {
		pn = args[3]
	}
	public, err := strconv.ParseBool(args[2])
	if err != nil {
		return err
	}

	s := NewCacheSrv(pn)

	for i := 0; i < nshard; i++ {
		if err := s.createShard(cache.Tshard(i), sp.NoFence(), make(Tcache)); err != nil {
			db.DFatalf("CreateShard %v\n", err)
		}
	}

	db.DPrintf(db.CACHESRV, "%v: Run %v\n", proc.GetName(), s.shrd)
	ssrv, err := sigmasrv.MakeSigmaSrvPublic(args[1]+s.shrd, s, db.CACHESRV, public)
	if err != nil {
		return err
	}
	if _, err := ssrv.Create(DUMP, sp.DMDIR|0777, sp.ORDWR, sp.NoLeaseId); err != nil {
		return err
	}
	if err := sessdevsrv.MkSessDev(ssrv.MemFs, DUMP, s.mkSession, nil); err != nil {
		return err
	}
	ssrv.RunServer()
	s.exitCacheSrv()
	return nil
}

func NewCacheSrv(pn string) *CacheSrv {
	cs := &CacheSrv{shards: make(map[cache.Tshard]*shardInfo), lastFence: sp.NullFence()}
	cs.tracer = tracing.Init("cache", proc.GetSigmaJaegerIP())
	p, err := perf.MakePerf(perf.CACHESRV)
	if err != nil {
		db.DFatalf("MakePerf err %v\n", err)
	}
	cs.perf = p
	cs.shrd = pn
	return cs
}

func (cs *CacheSrv) exitCacheSrv() {
	cs.tracer.Flush()
	cs.perf.Done()
}

//
// Fenced ops (with locking)
//

func (cs *CacheSrv) cmpFence(f sp.Tfence) sp.Tfencecmp {
	if !f.HasFence() {
		// cached runs without fence
		db.DPrintf(db.FENCEFS, "%v no fence %v\n", proc.GetName(), f)
	}
	if !cs.lastFence.IsInitialized() {
		db.DPrintf(db.FENCEFS, "%v initialize fence %v\n", proc.GetName(), f)
		cs.lastFence.Upgrade(&f)
		return sp.FENCE_EQ
	}
	return cs.lastFence.Cmp(&f)
}

func (cs *CacheSrv) cmpFenceUpgrade(f sp.Tfence) sp.Tfencecmp {
	cmp := cs.cmpFence(f)
	if cmp == sp.FENCE_LT {
		db.DPrintf(db.FENCEFS, "%v: New fence %v\n", proc.GetName(), f)
		cs.lastFence.Upgrade(&f)
	}
	return cmp
}

// For Put and Get
func (cs *CacheSrv) lookupShardFence(s cache.Tshard, f sp.Tfence) (*shardInfo, error) {
	cmp := cs.cmpFence(f)
	if cmp == sp.FENCE_LT {
		// cs is behind let the client retry until cs catches up
		db.DPrintf(db.ALWAYS, "%v: f %v shard %v cs behind; retry\n", proc.GetName(), cs.lastFence, s)
		return nil, serr.MkErr(serr.TErrRetry, fmt.Sprintf("shard %v", s))
	}
	sh, ok := cs.shards[s]
	if !ok {
		// if client is behind, return stale
		if cmp == sp.FENCE_GT {
			return nil, serr.MkErr(serr.TErrStale, fmt.Sprintf("shard %v", s))
		}
		// cs and client are in same config but server hasn't received
		// the shard yet.  let the client retry until the server
		// catchup first.
		db.DPrintf(db.ALWAYS, "%v: f %v shard %v cs waiting for shard; retry\n", proc.GetName(), cs.lastFence, s)
		return nil, serr.MkErr(serr.TErrRetry, fmt.Sprintf("shard %v", s))
	}
	switch sh.status {
	case READY:
		return sh, nil
	case FROZEN:
		return nil, serr.MkErr(serr.TErrStale, fmt.Sprintf("shard %v", s))
	default:
		db.DFatalf("lookupShardFence err status %v\n", sh.status)
		return nil, nil
	}
	return sh, nil
}

func (cs *CacheSrv) createShard(s cache.Tshard, f sp.Tfence, vals Tcache) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cmp := cs.cmpFenceUpgrade(f); cmp == sp.FENCE_GT {
		return serr.MkErr(serr.TErrStale, fmt.Sprintf("shard %v", s))
	}
	if _, ok := cs.shards[s]; ok {
		return serr.MkErr(serr.TErrExists, s)
	}
	sh := newShard()
	sh.fill(vals)
	cs.shards[s] = &shardInfo{status: READY, s: sh}
	return nil
}

func (cs *CacheSrv) CreateShard(ctx fs.CtxI, req cacheproto.ShardRequest, rep *cacheproto.CacheOK) error {
	db.DPrintf(db.CACHESRV, "CreateShard %v\n", req)
	return cs.createShard(req.Tshard(), req.Fence.Tfence(), req.Vals)
}

func (cs *CacheSrv) DeleteShard(ctx fs.CtxI, req cacheproto.ShardRequest, rep *cacheproto.CacheOK) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	db.DPrintf(db.CACHESRV, "DeleteShard %v\n", req)

	if _, ok := cs.shards[req.Tshard()]; !ok {
		return serr.MkErr(serr.TErrNotfound, req.Shard)
	}
	delete(cs.shards, req.Tshard())
	return nil
}

func (cs *CacheSrv) FreezeShard(ctx fs.CtxI, req cacheproto.ShardRequest, rep *cacheproto.CacheOK) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	db.DPrintf(db.CACHESRV, "FreezeShard %v\n", req)

	if cmp := cs.cmpFenceUpgrade(req.Fence.Tfence()); cmp == sp.FENCE_GT {
		return serr.MkErr(serr.TErrStale, fmt.Sprintf("shard %v", req.Tshard()))
	}
	si, ok := cs.shards[req.Tshard()]
	if !ok {
		return serr.MkErr(serr.TErrNotfound, req.Tshard())
	}
	switch si.status {
	case READY:
		si.status = FROZEN
	case FROZEN:
		db.DPrintf(db.ALWAYS, "%v: f %v %v already frozen\n", proc.GetName(), cs.lastFence, req.Tshard())
	}
	return nil
}

func (cs *CacheSrv) DumpShard(ctx fs.CtxI, req cacheproto.ShardRequest, rep *cacheproto.ShardData) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	db.DPrintf(db.CACHESRV, "DumpShard %v\n", req)

	if cmp := cs.cmpFence(req.Fence.Tfence()); cmp == sp.FENCE_GT {
		return serr.MkErr(serr.TErrStale, fmt.Sprintf("shard %v", req.Tshard()))
	}
	if si, ok := cs.shards[req.Tshard()]; !ok {
		return serr.MkErr(serr.TErrNotfound, req.Tshard())
	} else {
		rep.Vals = si.s.dump()
	}
	return nil
}

func (cs *CacheSrv) PutFence(ctx fs.CtxI, req cacheproto.CacheRequest, rep *cacheproto.CacheResult) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	db.DPrintf(db.CACHESRV, "PutFence %v\n", req)
	si, err := cs.lookupShardFence(req.Tshard(), req.Fence.Tfence())
	if err != nil {
		return err
	}
	if sp.Tmode(req.Mode) == sp.OAPPEND {
		err = si.s.append(req.Key, req.Value)
	} else {
		err = si.s.put(req.Key, req.Value)
	}
	return err
}

func (cs *CacheSrv) GetFence(ctx fs.CtxI, req cacheproto.CacheRequest, rep *cacheproto.CacheResult) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	db.DPrintf(db.CACHESRV, "GetFence %v\n", req)
	si, err := cs.lookupShardFence(req.Tshard(), req.Fence.Tfence())
	if err != nil {
		return err
	}
	v, ok := si.s.get(req.Key)
	if ok {
		rep.Value = v
		return nil
	}
	return serr.MkErr(serr.TErrNotfound, fmt.Sprintf("key %s", req.Key))
}

//
//  Unfenced ops and XXX locking
//

func (cs *CacheSrv) lookupShard(s cache.Tshard) (*shard, error) {
	sh, ok := cs.shards[s]
	if !ok {
		return nil, serr.MkErr(serr.TErrNotfound, fmt.Sprintf("shard %v", s))
	}
	if sh.status != READY {
		db.DFatalf("lookupShard %v err status %v\n", s, sh.status)
	}
	return sh.s, nil
}

func (cs *CacheSrv) Put(ctx fs.CtxI, req cacheproto.CacheRequest, rep *cacheproto.CacheResult) error {
	if req.Fence.HasFence() {
		return cs.PutFence(ctx, req, rep)
	}

	if false {
		_, span := cs.tracer.StartRPCSpan(&req, "Put")
		defer span.End()
	}

	db.DPrintf(db.CACHESRV, "Put %v\n", req)

	start := time.Now()

	s, err := cs.lookupShard(req.Tshard())

	if err != nil {
		return err
	}
	if sp.Tmode(req.Mode) == sp.OAPPEND {
		err = s.append(req.Key, req.Value)
	} else {
		err = s.put(req.Key, req.Value)
	}

	if time.Since(start) > 300*time.Microsecond {
		db.DPrintf(db.CACHE_LAT, "Long cache lock put: %v", time.Since(start))
	}
	return err
}

func (cs *CacheSrv) Get(ctx fs.CtxI, req cacheproto.CacheRequest, rep *cacheproto.CacheResult) error {
	if req.Fence.HasFence() {
		return cs.GetFence(ctx, req, rep)
	}

	e2e := time.Now()
	if false {
		_, span := cs.tracer.StartRPCSpan(&req, "Get")
		defer span.End()
	}

	db.DPrintf(db.CACHESRV, "Get %v", req)
	start := time.Now()

	s, err := cs.lookupShard(req.Tshard())
	if err != nil {
		return err
	}
	v, ok := s.get(req.Key)

	if time.Since(start) > 300*time.Microsecond {
		db.DPrintf(db.CACHE_LAT, "Long cache lock get: %v", time.Since(start))
	}

	if ok {
		rep.Value = v
		return nil
	}
	if time.Since(e2e) > 1*time.Millisecond {
		db.DPrintf(db.CACHE_LAT, "Long e2e get: %v", time.Since(e2e))
	}
	return serr.MkErr(serr.TErrNotfound, fmt.Sprintf("key %s", req.Key))
}

func (cs *CacheSrv) Delete(ctx fs.CtxI, req cacheproto.CacheRequest, rep *cacheproto.CacheResult) error {
	if false {
		_, span := cs.tracer.StartRPCSpan(&req, "Delete")
		defer span.End()
	}

	db.DPrintf(db.CACHESRV, "Delete %v", req)

	start := time.Now()

	s, err := cs.lookupShard(req.Tshard())
	if err != nil {
		return err
	}
	ok := s.delete(req.Key)

	if time.Since(start) > 20*time.Millisecond {
		db.DPrintf(db.ALWAYS, "Time spent witing for cache lock: %v", time.Since(start))
	}

	if ok {
		return nil
	}
	return serr.MkErr(serr.TErrNotfound, fmt.Sprintf("key %s", req.Key))
}
