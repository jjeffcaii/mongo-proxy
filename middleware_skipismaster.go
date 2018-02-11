package pxmgo

import (
	"time"

	"github.com/jjeffcaii/go-debug"
	"github.com/jjeffcaii/mongo-proxy/protocol"
)

type SkipIsMaster struct {
}

func (p *SkipIsMaster) Handle(ctx Context, req protocol.Message) error {
	q, ok := req.(*protocol.OpQuery)
	if !ok {
		return nil
	}
	if _, ok := protocol.Load(q.Query, "ismaster"); !ok {
		return nil
	}
	var doc = protocol.NewDocument().
		Set("ismaster", true).
		Set("maxBsonObjectSize", int32(16777216)).
		Set("maxMessageSizeBytes", int32(48000000)).
		Set("maxWriteBatchSize", int32(1000)).
		Set("localTime", time.Now().UnixNano()/1e6).
		Set("maxWireVersion", int32(5)).
		Set("minWireVersion", int32(0)).
		Set("readOnly", false).
		Set("ok", float64(1)).
		Build()
	if err := ctx.SendMessage(&protocol.OpReply{
		Op: &protocol.Op{
			OpHeader: &protocol.Header{
				OpCode: protocol.OpCodeReply,
			},
		},
		ResponseFlags:  8,
		CursorID:       0,
		NumberReturned: 1,
		Documents:      []protocol.Document{doc},
	}); err != nil {
		return err
	}
	debug.Debug("middware:skipismaster").Printf("skip is master: request_id=%d\n", q.OpHeader.RequestID)
	return Ignore
}
