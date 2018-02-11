package protocol

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var header = &Header{
	OpCode:        OpCodeQuery,
	ResponseTo:    1,
	RequestID:     2,
	MessageLength: 300,
}

func TestHeader_Encode(t *testing.T) {
	totals := 1000000
	start := time.Now().UnixNano()
	for i := 0; i < totals; i++ {
		_, err := header.Encode()
		if err != nil {
			t.Error(err)
		}
	}
	cost := time.Now().UnixNano() - start
	fmt.Println("encode ops:", 1e9*int64(totals)/cost)
}

func TestHeader_Decode(t *testing.T) {
	bs, err := header.Encode()
	if err != nil {
		t.Error(err)
	}
	header2 := &Header{}
	if err := header2.Decode(bs); err != nil {
		assert.Error(t, err, "decode header failed")
	}
	assert.Equal(t, header.OpCode, header2.OpCode, "decode OpCode failed")
	assert.Equal(t, header.ResponseTo, header2.ResponseTo, "decode ResponseTo failed")
	assert.Equal(t, header.RequestID, header2.RequestID, "decode RequestID failed")
	assert.Equal(t, header.MessageLength, header2.MessageLength, "decode MessageLength failed")
	totals := 1000000
	start := time.Now().UnixNano()
	for i := 0; i < totals; i++ {
		header2.Decode(bs)
	}
	cost := time.Now().UnixNano() - start
	fmt.Println("decode ops:", 1e9*int64(totals)/cost)
}
