package local

import (
	"context"
	"fmt"

	gsync "github.com/ipfs/go-graphsync"
	graphsync "github.com/ipfs/go-graphsync/impl"
	gsnet "github.com/ipfs/go-graphsync/network"

	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/libp2p/go-libp2p-core/host"
	dht "github.com/libp2p/go-libp2p-kad-dht"

	"github.com/0xPolygon/polygon-sdk/state"
	"github.com/anconprotocol/node/x/anconsync"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/libp2p/go-libp2p-core/peer"
)

// 	// event AddOnchainMetadata(string memory name, string memory description, string indexed memory image, string memory owner, string memory parent, bytes memory sources)

// 	// MetadataTransferOwnershipEvent represent the signature of
// 	// `event InitiateMetadataTransferOwnership(address fromOwner, address toOwner, string memory metadataUri)`
func MetadataTransferOwnershipEvent() abi.Event {

	addressType, _ := abi.NewType("address", "", nil)
	stringType, _ := abi.NewType("string", "", nil)
	return abi.NewEvent(
		"InitiateMetadataTransferOwnership",
		"InitiateMetadataTransferOwnership",
		false,
		abi.Arguments{abi.Argument{
			Name:    "fromOwner",
			Type:    addressType,
			Indexed: false,
		}, abi.Argument{
			Name:    "toOwner",
			Type:    addressType,
			Indexed: false,
		}, abi.Argument{
			Name:    "metadataUri",
			Type:    stringType,
			Indexed: false,
		}},
	)
}

func PostTxProcessing(s anconsync.Storage, t *state.Transition) error {
	for _, log := range t.Txn().Logs() {
		for _, topic := range log.Topics {
			if common.Hash(topic) != MetadataTransferOwnershipEvent().ID {
				continue
			}
		}
		if len(log.Topics) == 0 {
			continue
		}

		// if !ContractAllowed(log.Address) {
		// 	// Check the contract whitelist to prevent accidental native call.
		// 	continue
		// }
		values, err := MetadataTransferOwnershipEvent().Inputs.Unpack(log.Data)
		if err != nil {
			return err
		}
		fmt.Println(values...)

		if err != nil {
			continue
		}
		if err != nil {
			return err
		}
		break
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
