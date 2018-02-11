package pxmgo

func Pump(source Context, target Context) {
	go func(src Context, tgt Context) {
		ch1, ch2 := src.Next(), tgt.Next()
		for {
			var err error
			select {
			case msg := <-ch1:
				err = tgt.SendMessage(msg)
				break
			case msg := <-ch2:
				old := msg.Header().RequestID
				msg.Header().RequestID = 0
				var bs []byte
				bs, err = msg.Encode()
				msg.Header().RequestID = old
				if err == nil {
					err = src.Send(bs)
				}
				break
			}
			if err != nil {
				break
			}
		}
	}(source, target)
}
