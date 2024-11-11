package main

import (
	"crypto/rand"
	"fmt"

	"github.com/sean9999/gork"
)

func main() {

	var randy = rand.Reader
	alice := gork.NewPrincipal(randy, nil)
	//bob := gork.NewPrincipal(randy, nil)
	//eve := gork.NewPrincipal(randy, nil)
	//body := []byte("hello, world.")

	m := map[string]string{"cool": "beans"}
	g := gork.NewPrincipal(randy, m)
	gbytes, err := g.MarshalPEM()
	fmt.Println(err)
	fmt.Printf("%s\n\n", gbytes)
	p := g.AsPeer()
	pbytes, err := p.MarshalPEM()
	fmt.Println(err)
	fmt.Printf("%s\n\n", pbytes)

	//msg := alice.Compose(body, nil, bob.AsPeer())
	//msg.Recipient = bob.PublicKey()
	//msg.Encrypt(randy, &alice, nil)

	fmt.Println(alice.Art())
	//fmt.Println(bob.Art())

	fmt.Println(`
+---[RSA 3072]----+
|oo+o.++          |
|.o*o oo          |
| E.B.o.  ..      |
|. =.. =.+o.      |
| . +   XS=.      |
|. o   . *+.o     |
| . .    *.. .    |
|  o ...o + .     |
|   ..o.o+ .      |
+----[SHA256]-----+	
	`)

}
