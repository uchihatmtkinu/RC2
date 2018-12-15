package basic

import (
	"crypto/ecdsa"
	"math/big"
)

//RCSign is the signature type in our design
type RCSign struct {
	R *big.Int
	S *big.Int
}

//Miner is the miner
type Miner struct {
	ID        string
	Rep       int
	Prk       ecdsa.PublicKey
	LastGroup int
}

//OutType is the format of the output address data in the transaction
type OutType struct {
	Value   uint32
	Address [32]byte
}

//InType is the format of the input address data in the transaction
type InType struct {
	PrevTx [32]byte
	Index  uint32
	Sig    RCSign
	PukX   *big.Int
	PukY   *big.Int
}

//Transaction is the transaction data which sent by the sender
type Transaction struct {
	Timestamp uint64
	TxinCnt   uint32
	In        []InType
	TxoutCnt  uint32
	Out       []OutType
	Kind      uint32
	Locktime  uint32
	Hash      [32]byte
}

//TransactionBatch a set of multiple transactions
type TransactionBatch struct {
	TxCnt   uint32
	TxArray []Transaction
}

//TxList is the list of tx sent by Leader to miner for their verification
type TxList struct {
	ID       uint32
	HashID   [32]byte
	Round    uint32
	Level    uint32
	TxCnt    uint32
	TxArray  [][32]byte
	TxArrayX [][SHash]byte
	Sender   uint32
	Sig      RCSign
}

//TxDecision is the decisions based on given TxList
type TxDecision struct {
	ID       uint32   // miner id what id?  hash256(pubkey)
	HashID   [32]byte // transaction list id
	TxCnt    uint32
	Decision []byte
	Target   uint32
	Single   uint32
	Sig      []RCSign
}

//TxDecSet is the set of all decisions from one shard, signed by leader
type TxDecSet struct {
	ID         uint32
	Round      uint32
	Level      uint32
	HashID     [32]byte
	MemCnt     uint32
	ShardIndex uint32
	MemD       []TxDecision
	TxCnt      uint32
	TxArray    [][32]byte
	TxArrayX   [][SHash]byte
	Sender     uint32
	Sig        RCSign
}

//TxBlock introduce the struct of the transaction block
type TxBlock struct {
	ID            uint32
	Level         uint32
	PrevHash      [32]byte
	PrevFinalHash [32]byte
	ShardID       uint32
	HashID        [32]byte
	MerkleRoot    [32]byte
	Kind          uint32
	Timestamp     int64
	Height        uint32
	TxCnt         uint32
	TxArray       []Transaction
	TxHash        [][32]byte
	TxArrayX      [][SHash]byte
	Sender        uint32
	Sig           RCSign
}

//UserClient is the struct for miner and client
type UserClient struct {
	IPaddress string
	Prk       ecdsa.PublicKey
	kind      int
}

//AccCache is the cache of account
type AccCache struct {
	ID    [32]byte
	Value uint32
}
