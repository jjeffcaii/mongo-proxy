package pxmgo

import (
  "bufio"
  "bytes"
  "fmt"
  "log"
  "net"

  "github.com/jjeffcaii/go-mongo-proxy/protocol"
)

type mid struct {
  fn     Middleware
  allows []protocol.OpCode
}

func (p *mid) allow(req protocol.Message) bool {
  if p.allows == nil || len(p.allows) < 1 {
    return true
  }
  op := req.GetHeader().OpCode
  var found bool
  for _, v := range p.allows {
    if v == op {
      found = true
      break
    }
  }
  return found
}

type vContext struct {
  conn        net.Conn
  middlewares []*mid
  splicer     *splicer
  writer      *bufio.Writer
  reqId       int32
}

func (p *vContext) Use(middleware Middleware, allows ...protocol.OpCode) Context {
  if middleware != nil {
    p.middlewares = append(p.middlewares, &mid{
      fn:     middleware,
      allows: allows,
    })
  }
  return p
}

func (p *vContext) Next() (protocol.Message, error) {
  var bs []byte
  if data, err := p.splicer.next(); err != nil {
    return nil, err
  } else {
    bs = data.Bytes()
  }
  var msg protocol.Message
  opcode := protocol.ParseOpCode(bs)
  switch opcode {
  case protocol.OpCodeReply:
    msg = &protocol.OpReply{}
    break
  case protocol.OpCodeMsg:
    msg = &protocol.OpMsg{}
    break
  case protocol.OpCodeUpdate:
    msg = &protocol.OpUpdate{}
    break
  case protocol.OpCodeInsert:
    msg = &protocol.OpInsert{}
    break
  case protocol.RESERVED:
    // TODO: RESERVED
    break
  case protocol.OpCodeQuery:
    msg = &protocol.OpQuery{}
    break
  case protocol.OpCodeGetMore:
    msg = &protocol.OpGetMore{}
    break
  case protocol.OpCodeDel:
    msg = &protocol.OpDelete{}
    break
  case protocol.OpCodeKillCursor:
    msg = &protocol.OpKillCursors{}
    break
  case protocol.OpCodeCmd:
    msg = &protocol.OpCommand{}
    break
  case protocol.OpCodeCmdReply:
    msg = &protocol.OpCommandReply{}
    break
  default:
    break
  }
  if msg == nil {
    return nil, fmt.Errorf("bad message with opcode %d", opcode)
  }
  if err := msg.Decode(bs); err != nil {
    return nil, err
  }
  p.reqId = msg.GetHeader().RequestID
  return p.pipe(msg)
}

func (p *vContext) pipe(req protocol.Message) (protocol.Message, error) {
  l := len(p.middlewares)
  if l < 1 {
    return req, nil
  }
  ch := make(chan error)
  defer close(ch)
  first := p.middlewares[0]
  if first.allow(req) {
    go first.fn(req, p.sendMessage, p.genNext(req, 0, &ch))
  } else {
    go doNothing(req, p.sendMessage, p.genNext(req, 0, &ch))
  }
  err := <-ch
  switch err {
  case EOF:
    return req, nil
  case END:
    return p.Next()
  default:
    log.Println("middleware execute failed: ", err)
    return p.Next()
  }
}

func doNothing(_ protocol.Message, _ OnRes, next OnNext) {
  next(nil)
}

// 高阶函数, 用于生成next句柄.
func (p *vContext) genNext(req protocol.Message, index int, ch *chan error) func(error) {
  return func(err error) {
    if err != nil {
      *ch <- err
      return
    }
    i := index + 1
    if i >= len(p.middlewares) {
      *ch <- EOF
      return
    }
    mid := p.middlewares[i]
    if mid.allow(req) {
      mid.fn(req, p.sendMessage, p.genNext(req, i, ch))
    } else {
      doNothing(req, p.sendMessage, p.genNext(req, i, ch))
    }
  }
}

func (p *vContext) sendMessage(msg protocol.Message) error {
  old := msg.GetHeader().ResponseTo
  msg.GetHeader().ResponseTo = p.reqId
  bs, err := msg.Encode()
  msg.GetHeader().ResponseTo = old
  if err != nil {
    return err
  } else {
    return p.Send(bs)
  }
}

func (p *vContext) Send(bs []byte) error {
  _, err := p.writer.Write(bs)
  if err != nil {
    return err
  }
  return p.writer.Flush()
}

func (p *vContext) SendBuffer(bf *bytes.Buffer) error {
  _, err := bf.WriteTo(p.writer)
  if err != nil {
    return err
  }
  return p.writer.Flush()
}

func (p *vContext) Close() error {
  p.splicer.stop()
  return p.conn.Close()
}

func newContext(conn net.Conn) Context {
  return &vContext{
    conn:        conn,
    middlewares: make([]*mid, 0),
    splicer:     NewSplicer(bufio.NewReader(conn)),
    writer:      bufio.NewWriter(conn),
  }
}
