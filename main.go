package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/cosmos/go-bip39"

	"github.com/516108736/cosmos-tps/deploy"

	"github.com/516108736/cosmos-tps/query"

	"golang.org/x/sync/errgroup"

	"github.com/cosmos/cosmos-sdk/client/keys"
	cKyes "github.com/cosmos/cosmos-sdk/crypto/keys"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	cTypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/bank"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gaia/app"
)

func init() {
	var err error
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()
	// Instantiate the codec for the command line application
	cdc = app.MakeCodec()

	// Read in the configuration file for the sdk
	viper.Set(cli.HomeFlag, "/root/.gaiacli")
	viper.Set(client.FlagChainID, "mychain")
	viper.Set(client.FlagNode, "tcp://localhost:26657")
	viper.Set(flags.FlagTrustNode, true)
	viper.Set(flags.FlagBroadcastMode, flags.BroadcastSync)
	viper.Set(flags.FlagSkipConfirmation, true)

	keybase, err = keys.NewKeyBaseFromHomeFlag()
	checkErr(err)
}

var (
	password    = "11111111"
	cdc         = new(codec.Codec)
	keybase     = keys.NewInMemoryKeyBase()
	localCliCtx = new(Core)

	threadNum = 20
	preToNum  = 1000
	needTurns = 10

	typ          = flag.String("typ", "", "send or query")
	path         = flag.String("path", "", "send or query")
	sendInterval = flag.Int("i", 1, "send or query")
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

type Core struct {
	txBldr types.TxBuilder
	cliCtx context.CLIContext
}

func NewCoreFromAccress(from string, prepare bool) *Core {
	var err error
	txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
	cliCtx := context.NewCLIContextWithFrom(from).WithCodec(cdc)
	txBldr, err = utils.PrepareTxBuilder(txBldr, cliCtx)
	checkErr(err)

	return &Core{
		txBldr: txBldr,
		cliCtx: cliCtx,
	}
}

type BatchTxSend struct {
	LeaderAddress []string
	ToAddressList []string
	mmpp          map[string]*Core
	mpInfo        map[string]cKyes.Info
}

func NewBatchTxSend(leaderNum int, preTo int, config TXMAKE) *BatchTxSend {
	var err error
	b := &BatchTxSend{
		LeaderAddress: make([]string, leaderNum),
		ToAddressList: make([]string, preTo),
		mmpp:          make(map[string]*Core),
		mpInfo:        make(map[string]cKyes.Info),
	}

	for index := 0; index < leaderNum; index++ {
		name := fmt.Sprintf("validator%d", config.BaseIndex+index)
		b.LeaderAddress[index] = name
		b.mmpp[name] = NewCoreFromAccress(name, true)
		b.mpInfo[name], err = keybase.Get(name)
		checkErr(err)
	}
	fmt.Println("gen leader end", leaderNum, b.LeaderAddress)

	for index := 0; index < preTo; index++ {
		name := fmt.Sprintf("to%d", index+1)
		b.ToAddressList[index] = name
		b.mpInfo[name], err = keybase.Get(name)
		checkErr(err)
	}
	fmt.Println("gen follow end", preTo)
	return b
}

func WriteTx(txs []cTypes.Tx, txpath string) {
	bytes, err := json.Marshal(txs)
	checkErr(err)
	os.Remove(txpath)
	err = ioutil.WriteFile(txpath, bytes, 0644)
	checkErr(err)

}

func LoadTx(path string) []cTypes.Tx {
	context, err := ioutil.ReadFile(path)
	checkErr(err)
	txs := make([]cTypes.Tx, 0)
	err = json.Unmarshal(context, &txs)
	checkErr(err)
	return txs
}

func batchSend(c context.CLIContext, txs []cTypes.Tx) {
	start := 0
	interval := 1000
	flagEnd := 0
	for true {
		t := make([]cTypes.Tx, 0)
		if start+interval < len(txs) {
			t = txs[start : start+interval]
		} else {
			t = txs[start:len(txs)]
			flagEnd = 1
		}

		_, err := c.BroadcastTxSyncBatch(t)
		checkErr(err)

		if flagEnd == 1 {
			break
		}
		fmt.Println(time.Now().Format("2006/01/02 15:04:05"), "from", start, "to", start+interval)
		start = start + interval
		time.Sleep(time.Duration(*sendInterval) * time.Second)
	}
}

func send() {
	localCliCtx = NewCoreFromAccress("validator1", true)
	txPath := *path
	txs := LoadTx(txPath)
	batchSend(localCliCtx.cliCtx, txs)
}
func NewAccount(name string, password string) error {
	// read entropy seed straight from crypto.Rand and convert to mnemonic
	entropySeed, err := bip39.NewEntropy(256)
	if err != nil {
		return err
	}
	mnemonic, err := bip39.NewMnemonic(entropySeed[:])
	if err != nil {
		return err
	}
	kb, err := client.NewKeyBaseFromHomeFlag()
	if err != nil {
		return err
	}
	_, err = kb.CreateAccount(name, mnemonic, "", password, 0, 0)
	return err
}

func newAccount() {
	Len := 10000
	ts := time.Now()
	for index := 0; index <= Len; index++ {
		NewAccount(fmt.Sprintf("validator%d", index), "11111111")
		if index%1000 == 0 {
			fmt.Println("index-validator", index, time.Now().Sub(ts).Seconds())
		}
	}
	for index := 0; index <= Len; index++ {
		NewAccount(fmt.Sprintf("to%d", index), "11111111")
		if index%1000 == 0 {
			fmt.Println("index-to", index, time.Now().Sub(ts).Seconds())
		}
	}
	fmt.Println("newAccount end", time.Now().Sub(ts).Seconds(), Len)
}

func (b *BatchTxSend) makeTx(tb types.TxBuilder, clixtx context.CLIContext, from string, toAddr string, coinbase string, num int) [][]byte {
	toInfo := b.mpInfo[toAddr]
	coins, _ := sdk.ParseCoin(coinbase)
	fromAddr := clixtx.GetFromAddress()
	msg := bank.NewMsgSend(fromAddr, toInfo.GetAddress(), sdk.Coins{coins}, num)
	mm := make([]sdk.Msg, 0)
	for _, v := range msg {
		mm = append(mm, v)
	}
	tBytes, err := utils.CompleteTxCLIWithPassword(tb, clixtx, mm, password)
	checkErr(err)
	return tBytes
}

type TXMAKE struct {
	BaseIndex int
	Path      string
}

var (
	scfTxs = []TXMAKE{
		{
			BaseIndex: 100,
			Path:      "scfTx.json",
		},
		{
			BaseIndex: 200,
			Path:      "scfTx1.json",
		}, {
			BaseIndex: 300,
			Path:      "scfTx2.json",
		}, {
			BaseIndex: 400,
			Path:      "scfTx3.json",
		},
	}
)

func MakeTx() {
	localCliCtx = NewCoreFromAccress("validator1", true)
	ts := time.Now()
	for _, v := range scfTxs {
		makeTx(v)
	}
	fmt.Println("MakeTx end", time.Now().Sub(ts).Seconds())
}

func makeTx(config TXMAKE) {
	ts := time.Now()

	res := make(chan []cTypes.Tx, 0)
	rr := make([]cTypes.Tx, 0)

	b := NewBatchTxSend(threadNum, preToNum, config)

	var g errgroup.Group
	for _, leaderAddr := range b.LeaderAddress {
		fromAddr := leaderAddr
		g.Go(func() error {
			c := b.mmpp[fromAddr]
			for ii, toAddr := range b.ToAddressList {
				t := b.makeTx(c.txBldr.WithSequence(c.txBldr.Sequence()+uint64(needTurns)*uint64(ii)), c.cliCtx, fromAddr, toAddr, "1validatortoken", needTurns)
				txs := make([]cTypes.Tx, 0)
				for _, v := range t {
					txs = append(txs, v)
				}
				res <- txs
			}
			return nil
		})

	}

	for {
		select {
		case dd := <-res:
			rr = append(rr, dd...)
		}
		if len(rr) == len(b.LeaderAddress)*len(b.ToAddressList)*needTurns {
			break
		}
	}
	err := g.Wait()
	checkErr(err)
	WriteTx(rr, config.Path)

	fmt.Println("makeTx end", len(rr), config, time.Now().Sub(ts).Seconds())
}

func main() {
	flag.Parse()
	switch *typ {
	case "make":
		MakeTx()
	case "query":
		query.Start(10)
	case "send":
		send()
	case "deploy":
		deploy.Start()
	case "acc":
		newAccount()

	}
}
