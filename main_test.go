package gork

import (
	"crypto/rand"
	"encoding/pem"
	"os"
	"testing"

	"github.com/sean9999/go-delphi"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var realFs = afero.NewOsFs()
var lateSilencePath = "testdata/late-silence.config.json"

func writableConf(t testing.TB) afero.File {
	t.Helper()
	fd, err := realFs.OpenFile(lateSilencePath, os.O_RDWR|os.O_TRUNC, 0640)
	if err != nil {
		t.Error(err)
	}
	return fd
}

func readableConf(t testing.TB) afero.File {
	t.Helper()
	fd, err := realFs.Open(lateSilencePath)
	if err != nil {
		t.Error(err)
	}
	return fd
}

func TestNewGork(t *testing.T) {

	var randy = rand.Reader

	alice := ConstructPrincipal(WithRand(randy), WithConfig(readableConf(t)))

	alice.Props["hometown"] = "wonderland"

	bob := ConstructPrincipal(WithRand(randy), WithProps(map[string]string{
		"first_name": "bob",
		"age":        "47",
	}))

	eve := ConstructPrincipal(WithRand(randy))

	//eve := NewPrincipal(randy, nil, nil)
	eve.Props["lastName"] = "Macdonald"

	body := []byte("hello, world.")

	t.Run("encrypt / decrypt", func(t *testing.T) {
		msg := delphi.NewMessage(randy, []byte("hello world"))
		msg.Recipient = bob.PublicKey()
		err := alice.Encrypt(randy, msg, nil)
		assert.NoError(t, err)
		msg2 := alice.Compose(body, nil, bob.AsPeer())
		err = alice.Encrypt(randy, msg2, nil)
		assert.NoError(t, err)
		err = bob.Decrypt(msg2, nil)
		assert.NoError(t, err)
		assert.Equal(t, body, msg2.PlainText)
	})

	t.Run("sign / validate", func(t *testing.T) {
		msg := alice.Compose(body, nil, bob.AsPeer())
		err := msg.Sign(randy, alice)
		assert.NoError(t, err)
		valid := msg.Valid()
		assert.True(t, valid)
		digest, err := msg.Digest()
		assert.NoError(t, err)
		bob.Verify(msg.Sender, digest, msg.Signature())
	})

	t.Run("add / remove peer", func(t *testing.T) {
		err := alice.AddPeer(bob.AsPeer())
		assert.NoError(t, err)
		has := alice.HasPeer(bob.AsPeer())
		assert.True(t, has)
		err = alice.AddPeer(bob.AsPeer())
		assert.ErrorIs(t, err, ErrPeerExists)
		assert.Len(t, alice.Peers, 1)
		alice.DropPeer(eve.AsPeer())
		assert.Len(t, alice.Peers, 1)
		alice.DropPeer(bob.AsPeer())
		assert.Len(t, alice.Peers, 0)
	})

	t.Run("PEM encode / decode of principals", func(t *testing.T) {
		p, err := alice.MarshalPEM()
		assert.NoError(t, err)
		alice2 := new(Principal)
		err = alice2.UnmarshalPEM(p)
		assert.NoError(t, err)
		assert.Equal(t, alice.PublicKey().Bytes(), alice2.PublicKey().Bytes())
		assert.Equal(t, alice.PrivateKey().Bytes(), alice2.PrivateKey().Bytes())
		eq := alice.PrivateKey().Equal(alice2.PrivateKey())
		assert.True(t, eq)
		yuk := new(Principal)
		err = yuk.UnmarshalPEM([]byte("asdfasdfasdfasdfasdfasdfasdf"))
		assert.ErrorIs(t, err, ErrBadPem)
		emptyPem := pem.Block{
			Headers: map[string]string{},
		}
		b := pem.EncodeToMemory(&emptyPem)
		err = yuk.UnmarshalPEM(b)
		assert.ErrorIs(t, err, delphi.ErrBadKey)
		emptyPem.Headers["pubkey"] = "some invalid hex"
		b = pem.EncodeToMemory(&emptyPem)
		err = yuk.UnmarshalPEM(b)
		assert.ErrorIs(t, err, ErrBadPem)
	})

	t.Run("adding props and peers, exporting data", func(t *testing.T) {
		alice.Props["name"] = "Alice"
		err := alice.AddPeer(bob.AsPeer())
		assert.NoError(t, err)
		err = alice.Save(writableConf(t))
		assert.NoError(t, err)
	})

	// t.Run("validate signature", func(t *testing.T) {
	// 	prov := FileBasedConfigProvider{
	// 		Fs:   afero.NewOsFs(),
	// 		Name: "testdata/late-silence.config.json",
	// 	}
	// 	alice.WithConfigProvider(&prov)
	// })

}
