package middware

import "encoding/base64"

type stdGenerator struct{}

var gen = &stdGenerator{}

func (g *stdGenerator) GetNonce(ln int) []byte {
	if ln == 21 {
		return []byte("fyko+d2lbbFgONRv9qkxdawL") // Client's nonce
	}
	return []byte("3rfcNHYJY1ZVvWVs7j") // Server's nonce
}

func (g *stdGenerator) GetSalt(ln int) []byte {
	b, err := base64.StdEncoding.DecodeString("QSXCR+Q6sek8bf92")
	if err != nil {
		panic(err)
	}
	return b
}

func (g *stdGenerator) GetIterations() int {
	return 4096
}

