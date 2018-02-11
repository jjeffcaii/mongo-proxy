package protocol

type Op struct {
	OpHeader *Header
}

func (p *Op) Header() *Header {
	return p.OpHeader
}
