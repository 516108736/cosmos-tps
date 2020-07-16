package query

import (
	"fmt"
	"sync"
	"time"

	"github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client/context"
)

var (
	cliCtx = context.CLIContext{}
)

func initCli() {
	cliCtx = context.NewCLIContext()
}

func getBlock(h *int64) *types.Block {
	node, err := cliCtx.GetNode()
	checkErr(err)
	b, err := node.Block(h)
	checkErr(err)
	if b == nil {
		panic("b==nil")
	}
	return b.Block
}

type mmpp struct {
	mu         sync.RWMutex
	mp         map[int64]*types.Block
	interval   int
	highHeight int64
	lowHeight  int64
}

func NewBlockMap(interval int) *mmpp {
	b := getBlock(nil)

	return &mmpp{mp: map[int64]*types.Block{b.Height: b}, interval: interval, highHeight: b.Height, lowHeight: b.Height}

}

func (m *mmpp) GetTxLen(height int64) (int64, int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mp[height].Time.Unix(), len(m.mp[height].Txs)
}
func (m *mmpp) Set(height int64, b *types.Block) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mp[height] = b

}

func (m *mmpp) Check(from int64, to int64) {
	if from+1 == to {
		return
	}
	for i := from + 1; i < to; i++ {
		h := int64(i)
		b := getBlock(&h)
		m.Set(b.Height, b)
	}
}

func (m *mmpp) loop() {
	tickerLoop := time.NewTicker(1 * time.Second)
	defer tickerLoop.Stop()
	tickerDisplay := time.NewTicker(time.Duration(m.interval) * time.Second)
	defer tickerDisplay.Stop()
	for {
		select {
		case <-tickerLoop.C:
			b := getBlock(nil)
			m.Set(b.Height, b)
			m.Check(m.highHeight, b.Height)
			m.highHeight = b.Height
		case <-tickerDisplay.C:
			highHeightTime, _ := m.GetTxLen(m.highHeight)
			sum := 0
			lowTimeI := int64(0)
			detail := ""
			for index := m.highHeight; index >= m.lowHeight; index-- {
				ti, tLen := m.GetTxLen(index)

				sum += tLen
				detail += fmt.Sprintf("%d-", tLen)
				lowTimeI = ti
				if highHeightTime-ti > 60 {
					break
				}
			}
			fmt.Println(time.Now().Format("2006/01/02 15:04:05"), "intervalTime", highHeightTime-lowTimeI, "sumTx", sum, "tps", sum/60, "detail:", detail)
		}

	}

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func loopFind() {

}

func Start(interval int) {
	initCli()
	bm := NewBlockMap(interval)
	go bm.loop()
	select {}
}
