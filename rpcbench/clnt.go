package rpcbench

import (
	db "sigmaos/debug"
	"sigmaos/fslib"
	"sigmaos/rpcclnt"
	"sigmaos/rpcbench/proto"
	"sigmaos/sigmaclnt"
	"sigmaos/tracing"
)

type Clnt struct {
	c *rpcclnt.RPCClnt
	t *tracing.Tracer
}

func MakeClnt(sc *sigmaclnt.SigmaClnt, t *tracing.Tracer, path string) *Clnt {
	rpcc, err := rpcclnt.MkRPCClnt([]*fslib.FsLib{sc.FsLib}, path)
	if err != nil {
		db.DFatalf("Error MakeClnt: %v", err)
	}
	return &Clnt{
		c: rpcc,
		t: t,
	}
}

func (c *Clnt) Sleep(ms int64) error {
	_, span := c.t.StartTopLevelSpan("Clnt.Sleep")
	defer span.End()

	var res proto.SleepResult
	err := c.c.RPC("Srv.Sleep", &proto.SleepRequest{
		DurMS:             ms,
		SpanContextConfig: tracing.SpanToContext(span),
	}, &res)
	if err != nil {
		return err
	}
	return nil
}
