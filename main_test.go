package gork

import (
	"crypto/rand"
	"testing"

	"github.com/sean9999/go-delphi"
	"github.com/stretchr/testify/assert"
)

func TestNewGork(t *testing.T) {

	var randy = rand.Reader
	alice := NewGork(randy)
	bob := NewGork(randy)

	t.Run("encrypt / decrypt", func(t *testing.T) {
		msg := delphi.NewMessage(randy, []byte("hello world"))
		msg.Recipient = bob.PublicKey()
		err := alice.Encrypt(randy, msg, nil)
		assert.NoError(t, err)
	})

	t.Run("add / remove peer", func(t *testing.T) {
		err := alice.AddPeer(bob.AsPeer())
		assert.NoError(t, err)
		has := alice.HasPeer(bob.AsPeer())
		assert.True(t, has)
		err = alice.AddPeer(bob.AsPeer())
		assert.ErrorIs(t, err, ErrPeerExists)
		assert.Len(t, alice.Peers, 1)
		alice.DropPeer(bob.AsPeer())
		assert.Len(t, alice.Peers, 0)
	})

	t.Run("PEM encode / decode", func(t *testing.T) {
		pem, err := alice.MarshalPEM()
		assert.NoError(t, err)
		var alice2 Gork
		err = alice2.UnmarshalPEM(pem)
		assert.NoError(t, err)
		assert.Equal(t, alice.PublicKey().Bytes(), alice2.PublicKey().Bytes())
		assert.Equal(t, alice.PrivateKey().Bytes(), alice2.PrivateKey().Bytes())
		eq := alice.PrivateKey().Equal(alice2.PrivateKey())
		assert.True(t, eq)
	})

}
