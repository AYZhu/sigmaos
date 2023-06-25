package ctx

import (
	"encoding/json"

	db "sigmaos/debug"
	"sigmaos/sesscond"
	"sigmaos/sessp"
	sp "sigmaos/sigmap"
)

type CtxSnapshot struct {
	Uname  sp.Tuname
	Sessid sessp.Tsession
}

func MakeCtxSnapshot() *CtxSnapshot {
	cs := &CtxSnapshot{}
	return cs
}

func (ctx *Ctx) Snapshot() []byte {
	cs := MakeCtxSnapshot()
	cs.Uname = ctx.uname
	cs.Sessid = ctx.sessid
	b, err := json.Marshal(cs)
	if err != nil {
		db.DFatalf("Error snapshot encoding context: %v", err)
	}
	return b
}

func Restore(sct *sesscond.SessCondTable, b []byte) *Ctx {
	cs := MakeCtxSnapshot()
	err := json.Unmarshal(b, cs)
	if err != nil {
		db.DFatalf("error unmarshal ctx in restore: %v", err)
	}
	ctx := MkCtx(cs.Uname, cs.Sessid, sp.NoClntId, sct)
	return ctx
}
