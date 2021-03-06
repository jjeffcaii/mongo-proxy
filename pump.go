package pxmgo

import "io"

func Pump(source Context, target Context) {
	ch1, ch2 := source.Next(), target.Next()
	for {
		var err error
		select {
		case msg := <-ch1:
			if msg == nil {
				err = io.EOF
				break
			}
			err = target.SendMessage(msg)
			break
		case msg := <-ch2:
			if msg == nil {
				err = io.EOF
				break
			}
			old := msg.Header().RequestID
			msg.Header().RequestID = 0
			var bs []byte
			bs, err = msg.Encode()
			msg.Header().RequestID = old
			if err == nil {
				err = source.Send(bs)
			}
			break
		}
		if err != nil {
			break
		}
	}
}
