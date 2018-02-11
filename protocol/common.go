package protocol

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
