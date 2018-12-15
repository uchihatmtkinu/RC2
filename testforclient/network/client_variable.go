package network

import (
	"time"

	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/Reputation/cosi"
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/ed25519"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/rccache"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 16
const bufferSize = 1000
const timeoutCosi = 15 * time.Second //10seconds for timeout
const timeoutSync = 10 * time.Second
const timeSyncNotReadySleep = 5 * time.Second
const timeoutResponse = 120 * time.Second
const timeoutTL = 60 * time.Second
const timeoutNewRep = 15 * time.Second
const timeoutTxGap = 200 * time.Microsecond
const timeoutTxDecRev = 5 * time.Second
const timeoutResentTxmm = 2 * time.Second
const timeoutGetTx = time.Microsecond * 100

//CurrentEpoch epoch now
var CurrentEpoch int
var CurrentRepRound int

//LeaderAddr leader address
var LeaderAddr string

//MyGlobalID my global ID
var MyGlobalID int

//var AddrMapToInd map[string]int //ip+port
//var GroupMems []shard.MemShard
//GlobalAddrMapToInd
//var GlobalAddrMapToInd map[string]int

//CacheDbRef local database
var CacheDbRef rccache.DbRef

//------------------- shard process ----------------------
//readyInfo
type readyInfo struct {
	ID    int
	Epoch int
}

//readyMemberCh channel used in shard process, indicates the ready of the member for a new epoch
var readyMemberCh chan readyInfo

//readyLeaderCh channel used in shard process, indicates the ready of other shards for a new epoch
var readyLeaderCh chan readyInfo

//------------------- rolling process -------------------------
type rollingInfo struct {
	ID     uint32
	Epoch  uint32
	Leader uint32
}

//------------------- rep pow process -------------------------
//powInfo used in pow
type powInfo struct {
	ID    int
	Round int
	Hash  [32]byte
	Nonce int
}

//requetRepInfo used in pow
type requetRepInfo struct {
	ID    int
	Round int
}

//RepBlockRxInfo receive rep block
type RepBlockRxInfo struct {
	Round int
	Block Reputation.RepBlock
}

//RxRepBlockCh,
var RxRepBlockCh chan *Reputation.RepBlock

//------------------- cosi process -------------------------
//announceInfo used in cosi announce
type announceInfo struct {
	ID      int
	Message []byte
	Round   int
	Epoch   int
}

//commitInfo used in commitCh
type commitInfo struct {
	ID     int
	Commit cosi.Commitment
	Round  int
	Epoch  int
}

// challengeInfo challenge info
type challengeInfo struct {
	AggregatePublicKey ed25519.PublicKey
	AggregateCommit    cosi.Commitment
	Round              int
	Epoch              int
}

//responseInfo response info
type responseInfo struct {
	ID    int
	Sig   cosi.SignaturePart
	Round int
	Epoch int
}

//channel used in cosi
//cosiAnnounceCh cosi announcement channel
var cosiAnnounceCh chan announceInfo

//cosiCommitCh		cosi commitment channel
var cosiCommitCh chan commitInfo
var cosiChallengeCh chan challengeInfo
var cosiResponseCh chan responseInfo
var cosiSigCh chan responseInfo

//finalSignal
var finalSignal chan []byte

var startRep chan repInfo
var startTx chan int
var startSync chan bool
var CosiData map[int]cosi.SignaturePart

//syncSBInfo sync block info
type repInfo struct {
	Last  bool
	Hash  [][32]byte
	Rep   *[]int64
	Round int
}

//---------------------- sync process -------------
//syncSBInfo sync block info
type syncSBInfo struct {
	ID    int
	Block Reputation.SyncBlock
}

//syncTBInfo tx block info
type syncTBInfo struct {
	ID    int
	Block basic.TxBlock
}

//syncRequestInfo request sync
type syncRequestInfo struct {
	ID    int
	Round int
	Epoch int
}

//txDecRev request sync
type txDecRev struct {
	ID    uint32
	Round uint32
}

//TxBRequestInfo request txB
type TxBRequestInfo struct {
	Address string
	Shard   int
	Height  int
}

type syncNotReadyInfo struct {
	ID    int
	Epoch int
}

type TxBatchInfo struct {
	ID      uint32
	ShardID uint32
	Epoch   uint32
	Round   uint32
	Data    []byte
}

type QTL struct {
	ID    uint32
	Round uint32
}

type QTDS struct {
	ID    uint32
	Round uint32
}

type QTB struct {
	ID    uint32
	Round uint32
}

//channel used in sync
//syncCh
var syncSBCh [gVar.ShardCnt]chan syncSBInfo
var syncTBCh [gVar.ShardCnt]chan syncTBInfo
var syncNotReadyCh [gVar.ShardCnt]chan bool

//ShardDone flag determine whether the shard process is done
var ShardDone bool

//CoSiFlag flag determine the CoSi process has began
var CoSiFlag bool

//SyncFlag flag determine the Sync process has began
var SyncFlag bool

//IntialReadyCh channel used to indicate the process start
var IntialReadyCh chan bool
var ShardReadyCh chan bool
var CoSiReadyCh chan bool
var SyncReadyCh chan bool

var waitForFB chan bool

//FinalTxReadyCh whether the FB is done
var FinalTxReadyCh chan bool

var StartLastTxBlock chan int
var StartNewTxlist chan bool
var StartSendingTx chan bool

var TxDecRevChan [gVar.NumTxListPerEpoch]chan txDecRev
var TLChan [gVar.NumTxListPerEpoch]chan uint32
var RepFinishChan [gVar.NumberRepPerEpoch]chan bool

var TxBatchCache chan TxBatchInfo

//var StopGetTx chan bool

var txMCh [gVar.NumTxListPerEpoch]chan txDecRev

var bindAddress string

var BatchCache [gVar.NumTxListPerEpoch][]TxBatchInfo

var TDSChan [gVar.NumTxListPerEpoch]chan int
var TBChan [gVar.NumTxListPerEpoch]chan int
var TBBChan [gVar.NumTxListPerEpoch]chan int

//var StartSendTx chan bool
var rollingChannel chan rollingInfo
var VTDChannel chan rollingInfo
var rollingTxB chan []byte
var FBSent chan bool
var startDone bool
