rm -rf test-chain test-chain

go run main.go secrets init --data-dir test-chain

rm genesis.json



go run main.go dev --log-level debug  --dev-interval 5  --premine 0x32A21c1bB6E7C20F547e930b53dAC57f42cd25F6

