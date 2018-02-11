package protocol

import (
	"fmt"
	"io/ioutil"
	"testing"
)



func TestQuery(t *testing.T) {
	bs, _ := ioutil.ReadFile("/Users/caiweiwei/Desktop/opquery.bin")
	msg := &OpQuery{}
	if err := msg.Decode(bs); err != nil {
		t.Error(err)
	}
	fmt.Println("----------------------------------------")
	fmt.Printf("header:  %+v\n", *msg.OpHeader)
	fmt.Printf("message: %+v\n", msg)
	fmt.Printf("query: %+v\n", msg.Query)
	fmt.Println("----------------------------------------")
	bs2, err := msg.Encode()
	if err != nil {
		t.Error(err)
	}
	msg2 := &OpQuery{}
	err = msg2.Decode(bs2)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("----------------------------------------")
	fmt.Printf("header:  %+v\n", *msg2.OpHeader)
	fmt.Printf("message: %+v\n", msg2)
	fmt.Printf("query: %+v\n", msg2.Query)
	fmt.Println("----------------------------------------")
}

func TestReply(t *testing.T) {
	bs, _ := ioutil.ReadFile("/Users/caiweiwei/Desktop/op_reply.bin")
	msg := &OpReply{}
	if err := msg.Decode(bs); err != nil {
		t.Error(err)
	}

	fmt.Println("----------------------------------------")
	fmt.Printf("header:  %+v\n", *msg.OpHeader)
	fmt.Printf("message: %+v\n", msg)
	fmt.Printf("documents: %+v\n", msg.Documents[0])
	fmt.Println("----------------------------------------")

	bs2, _ := msg.Encode()
	fmt.Printf("bs1: %X\n", bs)
	fmt.Printf("bs2: %X\n", bs2)
	msg = &OpReply{}
	if err := msg.Decode(bs2); err != nil {
		t.Error(err)
	} else {
		fmt.Println("----------------------------------------")
		fmt.Printf("header:  %+v\n", *msg.OpHeader)
		fmt.Printf("message: %+v\n", msg)
		fmt.Printf("documents: %+v\n", msg.Documents[0])
		fmt.Println("----------------------------------------")
	}
}

func TestCommand(t *testing.T) {
	bs, _ := ioutil.ReadFile("/Users/caiweiwei/Desktop/op_command.bin")
	msg := &OpCommand{}
	if err := msg.Decode(bs); err != nil {
		t.Error(err)
	}
	fmt.Printf("header: %+v\n", *msg.OpHeader)
	fmt.Printf("message: %+v\n", msg)
	fmt.Printf("Metadata: %+v\n", msg.Metadata)
	bs2, _ := msg.Encode()

	fmt.Printf("bs1: %X\n", bs)
	fmt.Printf("bs2: %X\n", bs2)

	if len(bs) != len(bs2) {
		t.Errorf("bytes size doesn't match: before=%d, after=%d.", len(bs), len(bs2))
	}

}

func TestCommandReply(t *testing.T) {
	bs, _ := ioutil.ReadFile("/Users/caiweiwei/Desktop/op_commandreply.bin")
	msg := &OpCommandReply{}
	if err := msg.Decode(bs); err != nil {
		t.Error(err)
	}
	fmt.Printf("header: %+v\n", *msg.OpHeader)
	fmt.Printf("message: %+v\n", msg)
	fmt.Printf("metadata: %+v\n", msg.Metadata)
	fmt.Printf("reply: %+v\n", msg.CommandReply)
	bs2, _ := msg.Encode()

	fmt.Printf("bs1: %X\n", bs)
	fmt.Printf("bs2: %X\n", bs2)

	if len(bs) != len(bs2) {
		t.Errorf("bytes size doesn't match: before=%d, after=%d.", len(bs), len(bs2))
	}
}
