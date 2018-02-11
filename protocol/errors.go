package protocol

import "fmt"

var errHeaderLength = fmt.Errorf(fmt.Sprintf("at least %d bytes", HeaderLength))

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
