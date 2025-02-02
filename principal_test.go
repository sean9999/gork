package gork

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewKV(t *testing.T) {
	assert := assert.New(t)
	kv := NewKV()
	assert.NotNil(kv)
	assert.Equal(kv.Len(), 0)
	kv.Set("foo", "bar")
	assert.Equal(kv.Len(), 1)
}

func TestNewPrincipal(t *testing.T) {

	assert := assert.New(t)

	privHex := "5C2F587F8FA3F324412F1E033E2ECD8EB32D8E3F7703C38F5533EB11F7D9AD38583E7047EA1533D1B503E6FC40F4EA7BB7C5E5501FE442C311E60478776E464F71D77DA7683D812E0D01000E0829529A314E3BC4AD21637DCF7FEBD83F05A6F8DC90530FE62A10DF7FEC4DA2E99C3CECF486362701A8DA549EBE6FD8E61A6B69"
	hexR := strings.NewReader(privHex)
	e := hex.NewDecoder(hexR)

	p, err := PrincipalFrom(e)
	assert.NoError(err)

	nick := p.AsPeer().Nickname()

	if nick != "shy-pine" {
		t.Error(nick)
	}

	// p := NewPrincipal(rand.Reader, nil, nil)
	// want := []byte("asdf")
	// if !bytes.Equal(p.PrivateKey().Bytes(), want) {
	// 	t.Error(p.PrivateKey().ToHex())
	// }
}
