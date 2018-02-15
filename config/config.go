package config

import (
	"context"
	"crypto/rand"

	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"

	circuit "github.com/libp2p/go-libp2p-circuit"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	pnet "github.com/libp2p/go-libp2p-interface-pnet"
	metrics "github.com/libp2p/go-libp2p-metrics"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	swarm "github.com/libp2p/go-libp2p-swarm"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"
	ma "github.com/multiformats/go-multiaddr"
)

// Config describes a set of settings for a libp2p node
type Config struct {
	Transports         []TptC
	Muxers             []MsMuxC
	SecurityTransports []MsSecC
	ListenAddrs        []ma.Multiaddr
	PeerKey            crypto.PrivKey
	Peerstore          pstore.Peerstore
	Protector          pnet.Protector
	Reporter           metrics.Reporter
	Relay              bool
	RelayOpts          []circuit.RelayOpt
	Insecure           bool
}

func (cfg *Config) NewNode(ctx context.Context) (host.Host, error) {
	// If no key was given, generate a random 2048 bit RSA key
	privKey := cfg.PeerKey
	if cfg.PeerKey == nil {
		priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
		if err != nil {
			return nil, err
		}
		privKey = priv
	}

	// Obtain Peer ID from public key
	pid, err := peer.IDFromPublicKey(privKey.GetPublic())
	if err != nil {
		return nil, err
	}

	// Create a new blank peerstore if none was passed in
	ps := cfg.Peerstore
	if ps == nil {
		ps = pstore.NewPeerstore()
	}
	ps.AddPrivKey(pid, cfg.PeerKey)
	ps.AddPubKey(pid, cfg.PeerKey.GetPublic())

	swrm := swarm.NewSwarm(ctx, pid, ps, cfg.Reporter)

	// TODO: make host implementation configurable.
	h := bhost.New(swrm)

	upgrader := new(tptu.Upgrader)
	upgrader.Protector = cfg.Protector
	upgrader.Secure, err = makeSecurityTransport(h, cfg.SecurityTransports)
	if err != nil {
		h.Close()
		return nil, err
	}

	upgrader.Muxer, err = makeMuxer(h, cfg.Muxers)
	if err != nil {
		h.Close()
		return nil, err
	}

	tpts, err := makeTransports(h, upgrader, cfg.Transports)
	if err != nil {
		h.Close()
		return nil, err
	}
	for _, t := range tpts {
		err = swrm.AddTransport(t)
		if err != nil {
			h.Close()
			return nil, err
		}
	}

	if cfg.Relay {
		err := circuit.AddRelayTransport(ctx, h, upgrader, cfg.RelayOpts...)
		if err != nil {
			h.Close()
			return nil, err
		}
	}

	// TODO: This method succeeds if listening on one address succeeds. We
	// should probably fail if listening on *any* addr fails.
	if err := h.Network().Listen(cfg.ListenAddrs...); err != nil {
		h.Close()
		return nil, err
	}

	return h, nil
}
