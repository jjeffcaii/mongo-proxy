package middware

import (
  "time"

  "github.com/jjeffcaii/go-mongo-proxy"
  "github.com/jjeffcaii/go-mongo-proxy/protocol"
)

func SkipIsMaster(req protocol.Message, res pxmgo.OnRes, next pxmgo.OnNext) {
  q, _ := req.(*protocol.OpQuery)
  m := protocol.ToMap(q.Query)
  if m["ismaster"] == nil {
    next(nil)
    return
  }
  var doc = protocol.NewDocument().
    Set("ismaster", true).
    Set("maxBsonObjectSize", int32(16777216)).
    Set("maxMessageSizeBytes", int32(48000000)).
    Set("maxWriteBatchSize", int32(1000)).
    Set("localTime", time.Now().UnixNano()/1000000).
    Set("maxWireVersion", int32(5)).
    Set("minWireVersion", int32(0)).
    Set("readOnly", false).
    Set("ok", float64(1)).
    Build()
  var rep = &protocol.OpReply{
    Header: &protocol.Header{
      OpCode: protocol.OpCodeReply,
    },
    ResponseFlags:  8,
    CursorID:       0,
    NumberReturned: 1,
    Documents:      []protocol.Document{doc},
  }
  res(rep)
  next(pxmgo.END)
}
