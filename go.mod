module github.com/516108736/cosmos-tps

go 1.14

require (
	github.com/cosmos/cosmos-sdk v0.37.13
	github.com/cosmos/gaia v0.0.0-20200612180116-d00db033d861
	github.com/cosmos/go-bip39 v0.0.0-20180819234021-555e2067c45d
	github.com/ethereum/go-ethereum v1.9.15
	github.com/pkg/sftp v1.11.0
	github.com/spf13/cobra v1.0.0 // indirect
	github.com/spf13/viper v1.6.1
	github.com/tendermint/tendermint v0.32.12
	golang.org/x/crypto v0.0.0-20200311171314-f7b00557c8c4
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
)
replace github.com/cosmos/cosmos-sdk => /mnt/hgfs/GOPATH/src/github.com/cosmos/cosmos-sdk
replace github.com/tendermint/tendermint => /mnt/hgfs/GOPATH/src/github.com/tendermint/tendermint