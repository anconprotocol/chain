rm -rf test-chain test-chain-1

go run main.go secrets init --data-dir test-chain-1

go run main.go genesis --consensus ibft --ibft-validators-prefix-path test-chain-1 --premine 0x32A21c1bB6E7C20F547e930b53dAC57f42cd25F6

go run main.go server --data-dir ./test-chain-1 --chain genesis.json --grpc :10000 --libp2p :10001 --jsonrpc :10002 --seal --log-level debug

