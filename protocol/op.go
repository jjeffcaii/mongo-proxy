package protocol

type Op struct {
	Header *Header
}

func (p *Op) GetHeader() *Header {
	return p.Header
}
