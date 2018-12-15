package rccache

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"strconv"
	"strings"
	"sync"

	newrep "github.com/uchihatmtkinu/RC/NewRep"
	"github.com/uchihatmtkinu/RC/base58"
	"github.com/uchihatmtkinu/RC/gVar"

	"github.com/uchihatmtkinu/RC/basic"
)

const dbFilex = "TxBlockchain.db"

//UTXOBucket is the bucket of UTXO
const UTXOBucket = "UTXO"

//ACCBucket is the bucket of UTXO
const ACCBucket = "ACC"

//TBBucket is the bucket of ACC
const TBBucket = "TxBlock"

//FBBucket is the bucket of Final Blocks
const FBBucket = "FBTxBlock"

//TXBucket is the txbucket
const TXBucket = "TxBucket"

//byteCompare is the func used for string compare
func byteCompare(a, b interface{}) int {
	switch a.(type) {
	case *basic.AccCache:
		tmp1 := a.(*basic.AccCache).ID
		tmp2 := b.(*basic.AccCache).ID
		return bytes.Compare(tmp1[:], tmp2[:])
	default:
		return bytes.Compare([]byte(a.([]byte)), []byte(b.([]byte)))
	}

}

//DbRef is the structure stores the cache of a miner for the database
type DbRef struct {
	//const value
	Mu           sync.RWMutex
	ID           uint32
	DB           TxBlockChain
	HistoryShard []uint32
	prk          ecdsa.PrivateKey
	BandCnt      uint32
	Badness      bool

	TXCache   map[[32]byte]*CrossShardDec
	HashCache map[[basic.SHash]byte][][32]byte
	//WaitHashCache map[[basic.SHash]byte]WaitProcess

	ShardNum uint32

	//Leader
	TLSCache [gVar.NumTxListPerEpoch]*basic.TxList
	TDSCache [gVar.NumTxListPerEpoch]*basic.TxDecSet
	TBBCache [gVar.NumTxListPerEpoch]*basic.TxBlock
	TLIndex  map[[32]byte]*TLGroup
	Now      *TLGroup
	Ready    []basic.Transaction
	TxB      *basic.TxBlock
	FB       [gVar.ShardCnt]*basic.TxBlock

	TDSCnt      []int
	TDSNotReady int

	//Miner
	TLNow *basic.TxDecision

	TLRound    uint32
	PrevHeight uint32

	TBCache  *[][32]byte
	RepCache [gVar.NumTxListPerEpoch][gVar.ShardSize]int64

	Leader uint32
	//Statistic function
	TxCnt      uint32
	RepRound   uint32
	RepVote    [gVar.NumNewRep][gVar.ShardSize * gVar.ShardCnt]newrep.NewRep
	RepFirMsg  [gVar.NumNewRep][gVar.ShardSize]newrep.RepMsg
	RepSecMsg  [gVar.NumNewRep][gVar.ShardSize]newrep.RepSecMsg
	Rep        [gVar.NumNewRep][gVar.ShardSize * gVar.ShardCnt]uint64
	RepByz     [gVar.NumNewRep][gVar.ShardSize]bool
	RepFirSig  [gVar.NumNewRep][gVar.ShardSize]bool
	RepSecSig  [gVar.NumNewRep][gVar.ShardSize]bool
	GossipPool [gVar.ShardSize]int
	TLCheck    [gVar.NumTxListPerEpoch]bool
	TDSCheck   [gVar.NumTxListPerEpoch]bool
	TBCheck    [gVar.NumTxListPerEpoch]bool
	RepHash    [gVar.NumNewRep][32]byte
}

//TLGroup is the group of TL
type TLGroup struct {
	TLS [gVar.ShardCnt]basic.TxList
	TDS [gVar.ShardCnt]basic.TxDecSet
}

//PreStat is used in pre-defined request
type PreStat struct {
	Stat    int
	Valid   []int
	Channel chan bool
}

//WaitProcess is the current wait process
/*type WaitProcess struct {
	DataTB  []*basic.TxBlock
	StatTB  []*PreStat
	IDTB    []int
	DataTL  []*basic.TxList
	StatTL  []*PreStat
	IDTL    []int
	DataTDS []*basic.TxDecSet
	StatTDS []*PreStat
	IDTDS   []int
}*/

//Clear refresh the data for next epoch
func (d *DbRef) Clear() {
	d.Now = nil
	d.TLRound += 3
	d.BandCnt = 0
	d.TXCache = make(map[[32]byte]*CrossShardDec, 200000)
	d.HashCache = make(map[[basic.SHash]byte][][32]byte, 200000)
	//d.WaitHashCache = make(map[[basic.SHash]byte]WaitProcess, 200000)
	d.TLIndex = make(map[[32]byte]*TLGroup, 100)
	d.TxCnt = 0
	d.TDSCnt = make([]int, gVar.ShardCnt)
	d.TDSNotReady = int(gVar.ShardCnt)
	d.Ready = nil
	d.RepRound = 0
	for i := uint32(0); i < gVar.NumTxListPerEpoch; i++ {
		d.TLCheck[i] = false
		d.TDSCheck[i] = false
		d.TBCheck[i] = false
		d.TLSCache[i] = nil
		d.TDSCache[i] = nil
		d.TBBCache[i] = nil
	}
	for i := uint32(0); i < gVar.NumNewRep; i++ {
		for j := uint32(0); j < gVar.ShardSize; j++ {
			d.RepVote[i][j] = newrep.NewRep{j, 0}
		}
	}
	for i := uint32(0); i < gVar.NumOfTxForTest*gVar.NumTxListPerEpoch*10; i++ {
		d.Ready = append(d.Ready, *GenerateTx(1, 1, 1, 0, i))
	}
	if len(*d.TBCache) != 0 {
		fmt.Println("Miner", d.ID, "Cache clear: TBCache is not empty")
		for i := 0; i < len(*d.TBCache); i++ {
			fmt.Println("TBCache of", d.ID, "-", i, ":", base58.Encode((*d.TBCache)[i][:]))
		}
	}
}

//New is the initilization of DbRef
func (d *DbRef) New(x uint32, prk ecdsa.PrivateKey) {
	d.ID = x
	d.Badness = false
	d.prk = prk
	d.TxCnt = 0
	d.BandCnt = 0
	d.DB.CreateNewBlockchain(strings.Join([]string{strconv.Itoa(int(d.ID)), dbFilex}, ""))
	d.TXCache = make(map[[32]byte]*CrossShardDec, 200000)
	d.TxB = d.DB.LatestTxBlock()
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		d.FB[i] = d.DB.LatestFinalTxBlock(i)
	}
	d.TDSCnt = make([]int, gVar.ShardCnt)
	d.TDSNotReady = int(gVar.ShardCnt)
	d.HistoryShard = nil
	d.TLIndex = make(map[[32]byte]*TLGroup, 100)
	//d.WaitHashCache = make(map[[basic.SHash]byte]WaitProcess, 200000)
	for i := uint32(0); i < gVar.NumTxListPerEpoch; i++ {
		d.TLCheck[i] = false
		d.TDSCheck[i] = false
		d.TBCheck[i] = false
		d.TLSCache[i] = nil
		d.TDSCache[i] = nil
		d.TBBCache[i] = nil
	}
	for i := uint32(0); i < gVar.NumNewRep; i++ {
		for j := uint32(0); j < gVar.ShardSize; j++ {
			d.RepVote[i][j] = newrep.NewRep{j, 0}
		}
	}
	for i := uint32(0); i < gVar.NumOfTxForTest*gVar.NumTxListPerEpoch*10; i++ {
		d.Ready = append(d.Ready, *GenerateTx(1, 1, 1, 0, i))
	}
	d.HashCache = make(map[[basic.SHash]byte][][32]byte, 200000)
	d.TBCache = new([][32]byte)
	d.PrevHeight = 0

}

//CrossShardDec  is the database of cache
type CrossShardDec struct {
	Data     *basic.Transaction
	Decision [gVar.ShardSize]byte
	InCheck  []int //-1: Output related
	//0: unknown, Not-related; 1: Yes; 2: No; 3: Related-noresult
	ShardRelated []uint32
	Res          int8 //0: unknown; 1: Yes; -1: No; -2: Can be deleted
	InCheckSum   int
	Total        int
	Value        uint32
	Visible      bool
}

//Print output the crossshard information
func (c *CrossShardDec) Print() {
	fmt.Println("-----------CrossShardDec-------------")
	fmt.Println("Tx Hash: ", base58.Encode(c.Data.Hash[:]), "Value: ", c.Value)
	fmt.Println("ShardRelated: ", c.ShardRelated)
	fmt.Println("IncheckSum: ", c.InCheckSum, " Detail: ", c.InCheck)
	fmt.Println("Decision: ", c.Decision)
}

//ClearCache is to handle the TxCache of hash
func (d *DbRef) ClearCache(HashID [32]byte) error {
	tmp := basic.HashCut(HashID)
	delete(d.HashCache, tmp)
	delete(d.TXCache, HashID)
	return nil
}

//AddCache is to handle the TxCache of hash
func (d *DbRef) AddCache(HashID [32]byte) error {
	tmp := basic.HashCut(HashID)
	xxx, ok := d.HashCache[tmp]
	if !ok {
		d.HashCache[tmp] = [][32]byte{HashID}
	} else {
		if xxx[0] == HashID {
			//fmt.Println("Existing:", base58.Encode(HashID[:]))
			return fmt.Errorf("Existing")
		}
		fmt.Println("fuck!!!")
		d.HashCache[tmp] = append(xxx, HashID)
		for i := 0; i < len(xxx); i++ {
			fmt.Println("hashCache", i, ":", base58.Encode(xxx[i][:]))
		}
		fmt.Println(base58.Encode(HashID[:]))
	}
	return nil
}
