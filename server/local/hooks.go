package local

import (
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
	gsync "github.com/ipfs/go-graphsync"
	graphsync "github.com/ipfs/go-graphsync/impl"
	gsnet "github.com/ipfs/go-graphsync/network"

	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/libp2p/go-libp2p-core/host"
	dht "github.com/libp2p/go-libp2p-kad-dht"

	"github.com/0xPolygon/polygon-sdk/state"
	"github.com/anconprotocol/node/x/anconsync"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/libp2p/go-libp2p-core/peer"
)

func MetadataTransferOwnershipEvent() abi.Event {

	stringType, _ := abi.NewType("string", "", nil)
	return abi.NewEvent(
		"AddOnchainMetadata",
		"AddOnchainMetadata",
		false,
		abi.Arguments{abi.Argument{
			Name:    "name",
			Type:    stringType,
			Indexed: false,
		}, abi.Argument{
			Name:    "description",
			Type:    stringType,
			Indexed: false,
		}, abi.Argument{
			Name:    "image",
			Type:    stringType,
			Indexed: true,
		}, abi.Argument{
			Name:    "owner",
			Type:    stringType,
			Indexed: true,
		}, abi.Argument{
			Name:    "parent",
			Type:    stringType,
			Indexed: true,
		}, abi.Argument{
			Name:    "sources",
			Type:    stringType,
			Indexed: false,
		}},
	)
}

func EncodeDagCborEvent() abi.Event {

	str, _ := abi.NewType("string", "", nil)
	return abi.NewEvent(
		"EncodeDagCbor",
		"EncodeDagCbor",
		false,
		abi.Arguments{abi.Argument{
			Name:    "path",
			Type:    str,
			Indexed: false,
		}, abi.Argument{
			Name:    "hexdata",
			Type:    str,
			Indexed: false,
		}},
	)
}

func StoreDagBlockDoneEvent() abi.Event {

	str, _ := abi.NewType("string", "", nil)
	return abi.NewEvent(
		"StoreDagBlockDone",
		"StoreDagBlockDone",
		false,
		abi.Arguments{abi.Argument{
			Name:    "path",
			Type:    str,
			Indexed: false,
		}, abi.Argument{
			Name:    "cid",
			Type:    str,
			Indexed: true,
		}},
	)
}
func encodeDagCborBlock(inputs abi.Arguments, data []byte) (datamodel.Node, datamodel.Link, error) {

	props, err := inputs.Unpack(data)
	if err != nil {
		return nil, nil, err
	}

	///	path := props[0].(string)
	values := props[1].(string)

	n, _ := anconsync.Decode(basicnode.Prototype.Any, values)
	p := cidlink.LinkPrototype{cid.Prefix{
		Version:  1,
		Codec:    0x0129,
		MhType:   0x12, // sha2-256
		MhLength: 32,   // sha2-256 hash has a 32-byte sum.
	}}

	lnk := p.BuildLink(data)

	if err != nil {
		return nil, nil, err
	}

	return n, lnk, nil
}

func EncodeDagJsonEvent() abi.Event {

	str, _ := abi.NewType("string", "", nil)
	return abi.NewEvent(
		"EncodeDagJson",
		"EncodeDagJson",
		false,
		abi.Arguments{abi.Argument{
			Name:    "path",
			Type:    str,
			Indexed: false,
		}, abi.Argument{
			Name:    "hexdata",
			Type:    str,
			Indexed: false,
		}},
	)
}
func encodeDagJsonBlock(inputs abi.Arguments, data []byte) (datamodel.Node, datamodel.Link, error) {

	props, err := inputs.Unpack(data)
	if err != nil {
		return nil, nil, err
	}

	///	path := props[0].(string)
	values := props[1].(string)
	bz := common.Hex2Bytes(values)

	js := hexutil.Bytes{}
	js.UnmarshalJSON(bz)

	n, _ := anconsync.Decode(basicnode.Prototype.Any, string(js))
	p := cidlink.LinkPrototype{cid.Prefix{
		Version:  1,
		Codec:    0x0129,
		MhType:   0x12, // sha2-256
		MhLength: 32,   // sha2-256 hash has a 32-byte sum.
	}}

	lnk := p.BuildLink(data)

	if err != nil {
		return nil, nil, err
	}

	return n, lnk, nil
}

func PostTxProcessing(s anconsync.Storage, t *state.Transition) error {
	for _, log := range t.Txn().Logs() {
		for _, topic := range log.Topics {

			if len(log.Topics) == 0 {
				continue
			}

			var node datamodel.Node
			var lnk datamodel.Link
			var err error
			switch {
			case common.Hash(topic) == MetadataTransferOwnershipEvent().ID:

				break
			case common.Hash(topic) == EncodeDagJsonEvent().ID:
				node, lnk, err = encodeDagJsonBlock(EncodeDagJsonEvent().Inputs, log.Data)
				if err != nil {
					return err
				}

			default:
				break
			}
			// if !ContractAllowed(log.Address) {
			// 	// Check the contract whitelist to prevent accidental native call.
			// 	continue
			// }
			fmt.Println(lnk.String())
			fmt.Println(node)
			t.EmitLog(log.Address, log.Topics, []byte(lnk.String()))

			if err != nil {
				continue
			}
			if err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func GetHooks(s anconsync.Storage) func(t *state.Transition) {
	return func(t *state.Transition) {
		PostTxProcessing(s, t)
	}
}

func NewRouter(ctx context.Context, gsynchost host.Host, s anconsync.Storage) gsync.GraphExchange {

	var pi *peer.AddrInfo
	for _, addr := range dht.DefaultBootstrapPeers {
		pi, _ = peer.AddrInfoFromP2pAddr(addr)
		// We ignore errors as some bootstrap peers may be down
		// and that is fine.
		gsynchost.Connect(ctx, *pi)
	}
	network := gsnet.NewFromLibp2pHost(gsynchost)

	// Add Ancon fsstore
	exchange := graphsync.New(ctx, network, s.LinkSystem)

	// var receivedResponseData []byte
	// var receivedRequestData []byte

	exchange.RegisterIncomingResponseHook(
		func(p peer.ID, responseData gsync.ResponseData, hookActions gsync.IncomingResponseHookActions) {
			fmt.Println(responseData.Status().String(), responseData.RequestID())
		})

	exchange.RegisterIncomingRequestHook(func(p peer.ID, requestData gsync.RequestData, hookActions gsync.IncomingRequestHookActions) {
		// var has bool
		// receivedRequestData, has = requestData.Extension(td.extensionName)
		// if !has {
		// 	hookActions.TerminateWithError(errors.New("Missing extension"))
		// } else {
		// 	hookActions.SendExtensionData(td.extensionResponse)
		// }
		hookActions.ValidateRequest()
		hookActions.UseLinkTargetNodePrototypeChooser(basicnode.Chooser)
		fmt.Println(requestData.Root(), requestData.ID(), requestData.IsCancel())
	})
	finalResponseStatusChan := make(chan gsync.ResponseStatusCode, 1)
	exchange.RegisterCompletedResponseListener(func(p peer.ID, request gsync.RequestData, status gsync.ResponseStatusCode) {
		select {
		case finalResponseStatusChan <- status:
		default:
		}
	})

	return exchange
}
