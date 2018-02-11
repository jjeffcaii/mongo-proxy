package protocol

import "fmt"

type documentBuilder struct {
	pairs []Pair
}

func (p *documentBuilder) Set(key string, val interface{}) *documentBuilder {
	p.pairs = append(p.pairs, Pair{Key: key, Val: val,})
	return p
}

func (p *documentBuilder) Build() Document {
	return p.pairs
}

func NewDocument() *documentBuilder {
	return &documentBuilder{}
}

type errMessageLength struct {
	need, actually int
}

func (p *errMessageLength) Error() string {
	return fmt.Sprintf("broken message bytes: need=%d, actually=%d", p.need, p.actually)
}

type errMessageOffset struct {
	offset, totals int
}

func (p *errMessageOffset) Error() string {
	return fmt.Sprintf("broken message: read=%d, total=%d", p.offset, p.totals)
}
