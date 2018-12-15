package Reputation

import (
	"math"
	"sync"

	"github.com/uchihatmtkinu/RC/Reputation/cosi"
)

//maxNonce used in pow
var (
	maxNonce = math.MaxInt32
)

//difficulty difficulty in pow
const difficulty = 4

//RepPowInfo receive pow info
type RepPowInfo struct {
	ID    int
	Round int
	Nonce int
	Hash  [32]byte
}

//channel used in rep pow
//RepPowRxCh rx pow repblock from others
var RepPowRxCh chan RepPowInfo

//RepPowTxCh tx a pow repblock
var RepPowTxCh chan RepPowInfo

//RepPowRxValidate flag - validate the received repblock
var RepPowRxValidate chan RepPowInfo

//MyRepBlockChain my reputation blockchain
var MyRepBlockChain *RepBlockchain

//RepBlockIter an iterator on repblockchain
var RepBlockChainIter *RepBlockchainIterator

//SafeSyncBlock used in sync block
type SafeSyncBlock struct {
	Block *SyncBlock
	Epoch int
	Mu    sync.RWMutex
}

//SafeSyncBlock used in sync block
type SafeRepBlock struct {
	Block *RepBlock
	Round int
	Mu    sync.RWMutex
}

//CurrentSyncBlock current sync block
var CurrentSyncBlock SafeSyncBlock

//CurrentSyncBlock current sync block
var CurrentRepBlock SafeRepBlock

//CurrentCoSignature current cosinature
var CurrentCoSignature cosi.SignaturePart

//NonceMap map for nonce
var NonceMap map[int]int

//IDToNonce InShardID to nonce
var IDToNonce []int

//StartCalPoWAnnounce start flag
var StartCalPoWAnnounce chan bool
