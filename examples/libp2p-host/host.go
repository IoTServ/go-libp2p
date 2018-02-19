package main

import (
	"crypto/rand"
	"fmt"

	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
)

func main() {
	// To construct a simple host with all the default settings, just use `New`
	h, err := libp2p.New()
	if err != nil {
		panic(err)
	}
	defer h.Close()

	fmt.Printf("Hello World, my hosts ID is %s\n", h.ID())

	// If you want more control over the configuration, you can specify some
	// options to the constructor

	// Set your own keypair
	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		panic(err)
	}

	h2, err := libp2p.New(
		// Use your own created keypair
		libp2p.Identity(priv),

		// Set your own listen address
		// The config takes an array of addresses, specify as many as you want.
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/9000"),
	)
	if err != nil {
		panic(err)
	}
	defer h.Close()

	fmt.Printf("Hello World, my second hosts ID is %s\n", h2.ID())
}
