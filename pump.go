package pxmgo

import (
  "io"
  "log"
)

func Pump(source Context, target Context) {
  // frontend -> backend
  ch := make(chan int8)
  defer close(ch)
  go func(notifier chan<- int8) {
    defer func() {
      if e := recover(); e != nil {
        log.Println("something wrong before exiting: ", e)
      }
    }()
    for {
      m, err := source.Next()
      if err != nil {
        if err != io.EOF {
          log.Println("read from frontend failed: ", err)
        }
        break
      }
      bs, err := m.Encode()
      if err != nil {
        log.Println("encode source message failed: ", err)
        break
      }
      if err := target.Send(bs); err != nil {
        log.Println("send bytes to target failed: ", err)
        break
      }
    }
    notifier <- 0
  }(ch)
  // backend -> frontend
  go func(notifier chan<- int8) {
    if e := recover(); e != nil {
      log.Println("something wrong before exiting: ", e)
    }
    for {
      msg, err := target.Next()
      if err != nil {
        if err != io.EOF {
          log.Println("read from backend failed: ", err)
        }
        break
      }
      old := msg.GetHeader().RequestID
      msg.GetHeader().RequestID = 0
      bs, err := msg.Encode()
      msg.GetHeader().RequestID = old
      if err != nil {
        log.Println("encode target message failed: ", err)
        break
      }
      if err := source.Send(bs); err != nil {
        log.Println("send bytes to source failed: ", err)
        break
      }
    }
    notifier <- 1
  }(ch)
  <-ch
  if err := source.Close(); err != nil {
    log.Println("close source context failed: ", err)
  }
  if err := target.Close(); err != nil {
    log.Println("close target context failed: ", err)
  }
  <-ch
}
