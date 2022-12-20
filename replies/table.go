package replies

import (
	"fmt"
	"sort"
	"sync"

	"sigmaos/fcall"
	"sigmaos/intervals"
)

// Reply table for a given session.
type ReplyTable struct {
	sync.Mutex
	closed  bool
	entries map[fcall.Tseqno]*ReplyFuture
	// pruned has seqnos pruned from entries; client has received
	// the response for those.
	pruned *intervals.Intervals
}

func MakeReplyTable() *ReplyTable {
	rt := &ReplyTable{}
	rt.entries = make(map[fcall.Tseqno]*ReplyFuture)
	rt.pruned = intervals.MkIntervals()
	return rt
}

func (rt *ReplyTable) String() string {
	s := fmt.Sprintf("RT %d: ", len(rt.entries))
	keys := make([]fcall.Tseqno, 0, len(rt.entries))
	for k, _ := range rt.entries {
		keys = append(keys, k)
	}
	if len(keys) == 0 {
		return s + "\n"
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	start := keys[0]
	end := keys[0]
	for i := 1; i < len(keys); i++ {
		k := keys[i]
		if k != end+1 {
			s += fmt.Sprintf("[%d,%d) ", start, end)
			start = k
			end = k
		} else {
			end += 1
		}

	}
	s += fmt.Sprintf("[%d,%d)\n", start, end+1)
	return s
}

func (rt *ReplyTable) Register(request *fcall.FcallMsg) bool {
	rt.Lock()
	defer rt.Unlock()

	if rt.closed {
		return false
	}
	for s := request.Fc.Received.Start; s < request.Fc.Received.End; s++ {
		delete(rt.entries, fcall.Tseqno(s))
	}
	rt.pruned.Insert(request.Fc.Received)
	// if seqno in pruned, then drop
	if request.Fc.Seqno != 0 && rt.pruned.Contains(request.Fc.Seqno) {
		return false
	}
	rt.entries[request.Seqno()] = MakeReplyFuture()
	return true
}

// Expects that the request has already been registered.
func (rt *ReplyTable) Put(request *fcall.FcallMsg, reply *fcall.FcallMsg) bool {
	rt.Lock()
	defer rt.Unlock()

	s := request.Seqno()
	if rt.closed {
		return false
	}
	_, ok := rt.entries[s]
	if ok {
		rt.entries[s].Complete(reply)
	}
	return ok
}

func (rt *ReplyTable) Get(request *fcall.Fcall) (*ReplyFuture, bool) {
	rt.Lock()
	defer rt.Unlock()
	rf, ok := rt.entries[request.Tseqno()]
	return rf, ok
}

// Empty and permanently close the replies table. There may be server-side
// threads waiting on reply results, so make sure to complete all of them with
// an error.
func (rt *ReplyTable) Close(cli fcall.Tclient, sid fcall.Tsession) {
	rt.Lock()
	defer rt.Unlock()
	for _, rf := range rt.entries {
		rf.Abort(cli, sid)
	}
	rt.entries = make(map[fcall.Tseqno]*ReplyFuture)
	rt.closed = true
}

// Merge two reply caches.
func (rt *ReplyTable) Merge(rt2 *ReplyTable) {
	for seqno, entry := range rt2.entries {
		rf := MakeReplyFuture()
		if entry.reply != nil {
			rf.Complete(entry.reply)
		}
		rt.entries[seqno] = rf
	}
}
