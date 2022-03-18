package replraft

import (
	raft "go.etcd.io/etcd/raft/v3"

	np "ulambda/ninep"
	"ulambda/threadmgr"
)

type RaftReplServer struct {
	storage *raft.MemoryStorage
	node    *RaftNode
	clerk   *Clerk
}

func MakeRaftReplServer(id int, peerAddrs []string, tm *threadmgr.ThreadMgr) *RaftReplServer {
	srv := &RaftReplServer{}
	peers := []raft.Peer{}
	for i := range peerAddrs {
		peers = append(peers, raft.Peer{ID: uint64(i + 1)})
	}
	commitC := make(chan [][]byte)
	proposeC := make(chan []byte)
	srv.node = makeRaftNode(id, peers, peerAddrs, commitC, proposeC)
	srv.clerk = makeClerk(id, tm, commitC, proposeC)
	return srv
}

func (srv *RaftReplServer) Start() {
	go srv.clerk.serve()
}

func (srv *RaftReplServer) Process(fc *np.Fcall) {
	if fc.GetType() == np.TTdetach {
		msg := fc.Msg.(np.Tdetach)
		msg.PropId = uint32(srv.node.id)
		fc.Msg = msg
	}
	op := &Op{}
	op.request = fc
	op.reply = nil
	srv.clerk.request(op)
}
