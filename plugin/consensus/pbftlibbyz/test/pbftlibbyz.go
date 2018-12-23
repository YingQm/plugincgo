// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/33cn/chain33/blockchain"
	"github.com/33cn/chain33/common/limits"
	"github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/executor"
	"github.com/33cn/chain33/mempool"
	"github.com/33cn/chain33/p2p"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/rpc"
	"github.com/33cn/chain33/store"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/wallet"
	"github.com/33cn/plugincgo/plugin/consensus/pbftlibbyz"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugincgo/plugin/dapp/init"
	_ "github.com/33cn/plugincgo/plugin/store/init"
)

var (
	random *rand.Rand
	index  = flag.String("index", "1", "replica number")
)

func init() {
	err := limits.SetLimits()
	if err != nil {
		panic(err)
	}
	random = rand.New(rand.NewSource(types.Now().UnixNano()))
	log.SetLogLevel("info")
}

func main() {
	flag.Parse()

	var q = queue.New("channel")
	cfg, sub := types.InitCfg("chain33.test" + *index + ".toml")
	types.Init(cfg.Title, cfg)
	chain := blockchain.New(cfg.BlockChain)
	chain.SetQueueClient(q.Client())
	mem := mempool.New(cfg.Mempool, nil)
	mem.SetQueueClient(q.Client())
	exec := executor.New(cfg.Exec, sub.Exec)
	exec.SetQueueClient(q.Client())
	types.SetMinFee(0)
	s := store.New(cfg.Store, sub.Store)
	s.SetQueueClient(q.Client())
	cs := pbftlibbyz.Newpbftlibbyz(cfg.Consensus, sub.Consensus["pbftlibbyz"])
	cs.SetQueueClient(q.Client())
	p2pnet := p2p.New(cfg.P2P)
	p2pnet.SetQueueClient(q.Client())
	walletm := wallet.New(cfg.Wallet, sub.Wallet)
	walletm.SetQueueClient(q.Client())
	rpcapi := rpc.New(cfg.RPC)
	rpcapi.SetQueueClient(q.Client())

	defer chain.Close()
	defer mem.Close()
	defer p2pnet.Close()
	defer exec.Close()
	defer s.Close()
	defer cs.Close()
	defer q.Close()
	defer walletm.Close()
	defer rpcapi.Close()

	q.Start()
	clearTestData()
}

func clearTestData() {
	err := os.RemoveAll("datadir")
	if err != nil {
		fmt.Println("delete datadir have a err:", err.Error())
	}
	err = os.RemoveAll("wallet")
	if err != nil {
		fmt.Println("delete wallet have a err:", err.Error())
	}
	fmt.Println("test data clear successfully!")
}
