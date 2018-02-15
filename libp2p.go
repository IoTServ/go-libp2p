package libp2p

import (
	"context"
	"fmt"

	config "github.com/libp2p/go-libp2p/config"

	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	pnet "github.com/libp2p/go-libp2p-interface-pnet"
	metrics "github.com/libp2p/go-libp2p-metrics"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	secio "github.com/libp2p/go-libp2p-secio"
	ma "github.com/multiformats/go-multiaddr"
	mplex "github.com/whyrusleeping/go-smux-multiplex"
	yamux "github.com/whyrusleeping/go-smux-yamux"
)

type Option func(cfg *config.Config) error

// Chain chains multiple options into a single option.
func ChainOptions(opts ...Option) Option {
	return func(cfg *config.Config) error {
		for _, opt := range opts {
			if err := opt(cfg); err != nil {
				return err
			}
		}
		return nil
	}
}

func ListenAddrStrings(s ...string) Option {
	return func(cfg *config.Config) error {
		for _, addrstr := range s {
			a, err := ma.NewMultiaddr(addrstr)
			if err != nil {
				return err
			}
			cfg.ListenAddrs = append(cfg.ListenAddrs, a)
		}
		return nil
	}
}

func ListenAddrs(addrs ...ma.Multiaddr) Option {
	return func(cfg *config.Config) error {
		cfg.ListenAddrs = append(cfg.ListenAddrs, addrs...)
		return nil
	}
}

var DefaultSecurity Option = Security(secio.ID, secio.New)

var NoSecurity Option = func(cfg *config.Config) error {
	if len(cfg.SecurityTransports) > 0 {
		return fmt.Errorf("cannot use security transports with an insecure libp2p configuration")
	}
	cfg.Insecure = true
	return nil
}

func Security(name string, tpt interface{}) Option {
	return func(cfg *config.Config) error {
		if cfg.Insecure {
			return fmt.Errorf("cannot use security transports with an insecure libp2p configuration")
		}
		stpt, err := config.SecurityConstructor(tpt)
		if err == nil {
			cfg.SecurityTransports = append(cfg.SecurityTransports, config.MsSecC{stpt, name})
		}
		return err
	}
}

var DefaultMuxer Option = ChainOptions(
	Muxer("/yamux/1.0.0", yamux.DefaultTransport),
	Muxer("/mplex/6.3.0", mplex.DefaultTransport),
)

func Muxer(name string, tpt interface{}) Option {
	return func(cfg *config.Config) error {
		mtpt, err := config.MuxerConstructor(tpt)
		if err == nil {
			cfg.Muxers = append(cfg.Muxers, config.MsMuxC{mtpt, name})
		}
		return err
	}
}

func Transport(tpt interface{}) Option {
	return func(cfg *config.Config) error {
		tptc, err := config.TransportConstructor(tpt)
		if err == nil {
			cfg.Transports = append(cfg.Transports, tptc)
		}
		return err
	}
}

func Peerstore(ps pstore.Peerstore) Option {
	return func(cfg *config.Config) error {
		if cfg.Peerstore != nil {
			return fmt.Errorf("cannot specify multiple peerstore options")
		}

		cfg.Peerstore = ps
		return nil
	}
}

func PrivateNetwork(prot pnet.Protector) Option {
	return func(cfg *config.Config) error {
		if cfg.Protector != nil {
			return fmt.Errorf("cannot specify multiple private network options")
		}

		cfg.Protector = prot
		return nil
	}
}

func BandwidthReporter(rep metrics.Reporter) Option {
	return func(cfg *config.Config) error {
		if cfg.Reporter != nil {
			return fmt.Errorf("cannot specify multiple bandwidth reporter options")
		}

		cfg.Reporter = rep
		return nil
	}
}

func Identity(sk crypto.PrivKey) Option {
	return func(cfg *config.Config) error {
		if cfg.PeerKey != nil {
			return fmt.Errorf("cannot specify multiple identities")
		}

		cfg.PeerKey = sk
		return nil
	}
}

func New(ctx context.Context, opts ...Option) (host.Host, error) {
	var cfg config.Config
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}

	return cfg.NewNode(ctx)
}
