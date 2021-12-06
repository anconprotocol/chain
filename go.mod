module github.com/0xPolygon/polygon-sdk

go 1.14

require (
	github.com/anconprotocol/node v0.0.0-20211206142505-b3e9a67f7ada
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/ethereum/go-ethereum v1.10.13
	github.com/go-kit/kit v0.12.0
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/go-hclog v0.16.2
	github.com/hashicorp/go-immutable-radix v1.3.1
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/vault/api v1.3.0
	github.com/imdario/mergo v0.3.8
	github.com/ipfs/go-graphsync v0.9.3
	github.com/ipld/go-ipld-prime v0.14.0
	github.com/libp2p/go-libp2p v0.14.4
	github.com/libp2p/go-libp2p-core v0.8.6
	github.com/libp2p/go-libp2p-kad-dht v0.13.1
	github.com/libp2p/go-libp2p-kbucket v0.4.7
	github.com/libp2p/go-libp2p-noise v0.2.0
	github.com/libp2p/go-libp2p-pubsub v0.4.1
	github.com/mitchellh/cli v1.1.0
	github.com/multiformats/go-multiaddr v0.3.3
	github.com/prometheus/client_golang v1.11.0
	github.com/ryanuber/columnize v2.1.2+incompatible
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	github.com/umbracle/fastrlp v0.0.0-20210128110402-41364ca56ca8
	github.com/umbracle/go-eth-bn256 v0.0.0-20190607160430-b36caf4e0f6b
	github.com/umbracle/go-web3 v0.0.0-20210427125755-b411b4019e22
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	google.golang.org/grpc v1.42.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
