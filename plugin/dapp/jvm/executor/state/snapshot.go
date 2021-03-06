package state

import (
	"sort"

	"github.com/33cn/chain33/types"
	jvmtypes "github.com/33cn/plugincgo/plugin/dapp/jvm/types"
)

// DataChange 数据状态变更接口
// 所有的数据状态变更事件实现此接口，并且封装各自的变更数据以及回滚动作
// 在调用合约时（具体的Tx执行时），会根据操作生成对应的变更对象并缓存下来
// 如果合约执行出错，会按生成顺序的倒序，依次调用变更对象的回滚接口进行数据回滚，并同步删除变更对象缓存
// 如果合约执行成功，会按生成顺序的郑旭，依次调用变更对象的数据和日志变更记录，回传给区块链
type DataChange interface {
	getData(mdb *MemoryStateDB) []*types.KeyValue
	getLog(mdb *MemoryStateDB) []*types.ReceiptLog
}

// Snapshot 版本结构，包含版本号以及当前版本包含的变更对象在变更序列中的开始序号
type Snapshot struct {
	id      int
	entries []DataChange
	statedb *MemoryStateDB
}

// GetID get id for snapshot
func (ver *Snapshot) GetID() int {
	return ver.id
}

// 添加变更数据
func (ver *Snapshot) append(entry DataChange) {
	ver.entries = append(ver.entries, entry)
}

// 获取当前版本变更数据
func (ver *Snapshot) getData() (kvSet []*types.KeyValue, logs []*types.ReceiptLog) {
	// 获取中间的数据变更
	dataMap := make(map[string]*types.KeyValue)

	for _, entry := range ver.entries {

		items := entry.getData(ver.statedb)
		logEntry := entry.getLog(ver.statedb)
		if logEntry != nil {
			logs = append(logs, entry.getLog(ver.statedb)...)
		}

		// 执行去重操作
		for _, kv := range items {
			dataMap[string(kv.Key)] = kv
		}
	}

	// 因为需要进行去重处理，使用了容器map，
	// 这里可能会引起数据顺序不一致的问题，从而导致共识共识过程中的hash不一致问题，
	// 解决办法就是先list所有的name，然后再排序
	names := make([]string, 0, len(dataMap))
	for name := range dataMap {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		kvSet = append(kvSet, dataMap[name])
	}

	return kvSet, logs
}

type (

	// 基础变更对象，用于封装默认操作
	baseChange struct {
	}

	// 创建合约对象变更事件
	createAccountChange struct {
		baseChange
		account string
	}

	// 存储状态变更事件
	storageChange struct {
		baseChange
		account  string
		key      []byte
		prevalue []byte
	}

	// 本地存储状态变更事件
	localStorageChange struct {
		baseChange
		account  string
		key      []byte
		data     []byte
		prevalue []byte
	}

	// 合约代码状态变更事件
	codeChange struct {
		baseChange
		account  string
		prevcode []byte
		prevhash []byte
		prevabi  []byte
	}

	// 转账事件
	// 合约转账动作不执行回滚，失败后数据不会写入区块
	balanceChange struct {
		baseChange
		amount int64
		data   []*types.KeyValue
		logs   []*types.ReceiptLog
	}
)

func (ch baseChange) getData(s *MemoryStateDB) (kvset []*types.KeyValue) {
	return nil
}

func (ch baseChange) getLog(s *MemoryStateDB) (logs []*types.ReceiptLog) {
	return nil
}

// 创建账户对象的数据集
func (ch createAccountChange) getData(s *MemoryStateDB) (kvset []*types.KeyValue) {
	acc := s.accounts[ch.account]
	if acc != nil {
		kvset = append(kvset, acc.GetDataKV()...)
		return kvset
	}
	return nil
}

func (ch codeChange) getData(mdb *MemoryStateDB) (kvset []*types.KeyValue) {
	acc := mdb.accounts[ch.account]
	if acc != nil {
		kvset = append(kvset, acc.GetDataKV()...)
		return kvset
	}
	return nil
}

func (ch storageChange) getData(mdb *MemoryStateDB) []*types.KeyValue {
	value := mdb.GetState(ch.account, string(ch.key))
	if value == nil {
		return nil
	}
	acc := mdb.GetAccount(ch.account)
	key := acc.GetStateItemKey(ch.account, string(ch.key))

	return []*types.KeyValue{{Key: []byte(key), Value: value}}
}

func (ch storageChange) getLog(mdb *MemoryStateDB) []*types.ReceiptLog {
	return nil
}

func (ch balanceChange) getData(mdb *MemoryStateDB) []*types.KeyValue {
	return ch.data
}
func (ch balanceChange) getLog(mdb *MemoryStateDB) []*types.ReceiptLog {
	return ch.logs
}

func (ch localStorageChange) getData(mdb *MemoryStateDB) []*types.KeyValue {
	return nil
}

func (ch localStorageChange) getLog(mdb *MemoryStateDB) []*types.ReceiptLog {
	localData := &jvmtypes.ReceiptLocalData{
		Key:      ch.key,
		CurValue: ch.data,
		PreValue: ch.prevalue,
	}

	log := &types.ReceiptLog{
		Ty:  jvmtypes.TyLogLocalDataJvm,
		Log: types.Encode(localData),
	}

	return []*types.ReceiptLog{log}
}
