package npclnt

import (
	"sync"

	db "ulambda/debug"
	np "ulambda/ninep"
)

// XXX duplicate
const (
	Msglen = 64 * 1024
)

type ChanMgr struct {
	mu      sync.Mutex
	name    string
	session np.Tsession
	seqno   *np.Tseqno
	conns   map[string]*Chan
}

func makeChanMgr(session np.Tsession, seqno *np.Tseqno) *ChanMgr {
	cm := &ChanMgr{}
	cm.conns = make(map[string]*Chan)
	cm.session = session
	cm.seqno = seqno
	return cm
}

func (cm *ChanMgr) exit() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for addr, conn := range cm.conns {
		db.DLPrintf("9PCHAN", "exit close connection to %v\n", addr)
		conn.Close()
		delete(cm.conns, addr)
	}
}

// XXX Make array
func (cm *ChanMgr) allocChan(addr string) (*Chan, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var err error
	conn, ok := cm.conns[addr]
	if !ok {
		conn, err = mkChan([]string{addr})
		if err == nil {
			cm.conns[addr] = conn
		}
	}
	return conn, err
}

func (cm *ChanMgr) makeCall(dst string, req np.Tmsg) (np.Tmsg, error) {
	conn, err := cm.allocChan(dst)
	if err != nil {
		return nil, err
	}
	reqfc := &np.Fcall{}
	reqfc.Type = req.Type()
	reqfc.Msg = req
	reqfc.Session = cm.session
	reqfc.Seqno = cm.seqno.Next()
	repfc, err := conn.RPC(reqfc)
	if err != nil {
		return nil, err
	}
	return repfc.Msg, nil
}
