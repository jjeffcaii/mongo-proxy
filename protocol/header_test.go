package protocol

import (
	"fmt"
	"testing"
	"time"
)

func TestHeader_Encode(t *testing.T) {
	header := &Header{
		OpCode:        OpCodeQuery,
		ResponseTo:    1,
		RequestID:     2,
		MessageLength: 300,
	}
	totals := int64(1000000)
	start := time.Now().UnixNano()
	for i := int64(0); i < totals; i++ {
		_, err := header.Encode()
		if err != nil {
			t.Error(err)
		}
	}
	cost := time.Now().UnixNano() - start
	fmt.Println("ops:", 1e9*totals/cost)
}

func TestHeaderCodec(t *testing.T) {
	header := Header{
		OpCode:        OpCodeQuery,
		ResponseTo:    1,
		RequestID:     2,
		MessageLength: 300,
	}
	if bs, err := header.Encode(); err == nil {
		fmt.Printf("bytes: %X\n", bs)
		header2 := &Header{}
		header2.Decode(bs)
		fmt.Printf("header1: %+v\n", header)
		fmt.Printf("header2: %+v\n", header2)
		if header.OpCode != header2.OpCode {
			t.Error("check OpCode failed.")
		}
		if header.ResponseTo != header2.ResponseTo {
			t.Errorf("check ResponseTo failed.")
		}
		if header.RequestID != header2.RequestID {
			t.Errorf("check RequestID failed.")
		}
		if header.MessageLength != header2.MessageLength {
			t.Errorf("check MessageLength failed.")
		}
	} else {
		t.Error(err)
	}
}
