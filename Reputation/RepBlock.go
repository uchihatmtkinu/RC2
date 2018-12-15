package Reputation

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"

	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"

	"fmt"
	"github.com/uchihatmtkinu/RC/Reputation/cosi"
	"github.com/uchihatmtkinu/RC/ed25519"
)

// RepBlock reputation block
type RepBlock struct {
	Timestamp         int64
	RepTransactions   []*RepTransaction
	StartBlock        bool
	PrevSyncBlockHash [][32]byte
	PrevTxBlockHashes [][32]byte
	PrevRepBlockHash  [32]byte
	Hash              [32]byte
	Nonce             int
	Cosig			  cosi.SignaturePart
}

//NewRepBlock creates and returns Block
func NewRepBlock(repData *[]int64, startBlock bool, prevSyncRepBlockHash [][32]byte, prevTxBlockHashes [][32]byte, prevRepBlockHash [32]byte) *RepBlock {
	//var item *shard.MemShard
	var repTransactions []*RepTransaction
	tmpprevSyncRepBlockHash := make([][32]byte, len(prevSyncRepBlockHash))
	copy(tmpprevSyncRepBlockHash, prevSyncRepBlockHash)

	tmpprevTxBlockHashes := make([][32]byte, len(prevTxBlockHashes))
	copy(tmpprevTxBlockHashes, prevTxBlockHashes)
	for i := uint32(0); i < gVar.ShardSize; i++ {
		//item = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		repTransactions = append(repTransactions, NewRepTransaction(shard.ShardToGlobal[shard.MyMenShard.Shard][i], (*repData)[i]))
	}
	var block *RepBlock
	//generate new block
	if startBlock {
		block = &RepBlock{time.Now().Unix(), repTransactions, startBlock, tmpprevSyncRepBlockHash, tmpprevTxBlockHashes, [32]byte{gVar.MagicNumber}, [32]byte{}, 0,[]byte{0}}
	} else {
		block = &RepBlock{time.Now().Unix(), repTransactions, startBlock, nil, tmpprevTxBlockHashes, prevRepBlockHash, [32]byte{}, 0,[]byte{0}}
	}
	block.Nonce = 0
	data := block.prepareData()
	hash := sha256.Sum256(data)
	block.Hash = hash

	return block
}

// NewGenesisRepBlock creates and returns genesis Block
func NewGenesisRepBlock() *RepBlock {
	block := &RepBlock{time.Now().Unix(), nil, true, [][32]byte{{gVar.MagicNumber}}, [][32]byte{{gVar.MagicNumber}}, [32]byte{gVar.MagicNumber}, [32]byte{gVar.MagicNumber}, int(gVar.MagicNumber), []byte{gVar.MagicNumber}}
	return block
}

// HashRep returns a hash of the RepValue in the block
func (b *RepBlock) HashRep() []byte {
	var txHashes []byte
	var txHash [32]byte
	for _, item := range b.RepTransactions {
		txHashes = append(txHashes, IntToHex(item.Rep)[:]...)
		txHashes = append(txHashes, IntToHex(int64(item.GlobalID))...)
	}
	txHash = sha256.Sum256(txHashes)
	return txHash[:]
}

// HashPrevTxBlockHashes returns a hash of the prevTxBlocks in the block
func (b *RepBlock) HashPrevTxBlockHashes() []byte {
	var txHashes []byte
	var txHash [32]byte
	for _, item := range b.PrevTxBlockHashes {
		txHashes = append(txHashes, item[:]...)
	}
	txHash = sha256.Sum256(txHashes)
	return txHash[:]
}

// Serialize encode block
func (b *RepBlock) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

//Print is output
func (b *RepBlock) Print() {
	fmt.Println("RepBlock:")
	fmt.Println("RepTransactions:")
	for _, item := range b.RepTransactions {
		fmt.Print("	GlobalID:", item.GlobalID)
		fmt.Println("		Rep", item.Rep)
	}
	fmt.Println("StartBlock:", b.StartBlock)
	fmt.Println("PrevTxBlockHashes:", b.PrevTxBlockHashes)
	fmt.Println("PrevSyncRepBlockHash", b.PrevSyncBlockHash)
	fmt.Println("PrevRepBlockHash:", b.PrevRepBlockHash)
	fmt.Println("Hash:", b.Hash)

}


// VerifyCosign verify CoSignature
func (b *RepBlock) VerifyCoSignature(ms *[]shard.MemShard) bool {
	//verify signature
	var pubKeys []ed25519.PublicKey
	sbMessage := b.PrevRepBlockHash[:]
	pubKeys = make([]ed25519.PublicKey, int(gVar.ShardSize))
	for i,it:= range b.RepTransactions{
		pubKeys[i] = (*ms)[it.GlobalID].CosiPub
	}
	valid := cosi.Verify(pubKeys, cosi.ThresholdPolicy(int(gVar.ShardSize)/2), sbMessage, b.Cosig)
	return valid
}

func (b *RepBlock) prepareData() []byte {
	data := bytes.Join(
		[][]byte{
			b.HashRep(),
			b.HashPrevTxBlockHashes(),
			BoolToHex(b.StartBlock),
			b.PrevRepBlockHash[:],
			//IntToHex(pow.RepBlock.Timestamp),
			IntToHex(int64(b.Nonce)),
		},
		[]byte{},
	)

	return data
}



// DeserializeRepBlock decode Repblock
func DeserializeRepBlock(d []byte) *RepBlock {
	var block RepBlock
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}
