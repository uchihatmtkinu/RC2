package Reputation

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/uchihatmtkinu/RC/shard"
)

// ProofOfWork represents a proof-of-work
type ProofOfWork struct {
	RepBlock *RepBlock
	Target   *big.Int
}

// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWork(b *RepBlock) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-difficulty))

	pow := &ProofOfWork{b, target}

	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.RepBlock.HashRep(),
			pow.RepBlock.HashPrevTxBlockHashes(),
			BoolToHex(pow.RepBlock.StartBlock),
			pow.RepBlock.PrevRepBlockHash[:],
			//IntToHex(pow.RepBlock.Timestamp),
			IntToHex(int64(difficulty)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

// Run performs a proof-of-work
func (pow *ProofOfWork) Run() (int, [32]byte, bool) {
	var hashInt big.Int
	var hash [32]byte
	var flag bool
	var candidateRepBlock RepPowInfo
	flag = true
	nonce := 0
	fmt.Println("Mining the RepBlock containing")
	for nonce < maxNonce && flag {
		select {
		case candidateRepBlock = <-RepPowRxCh:
			{
				if candidateRepBlock.Round > CurrentRepBlock.Round {
					NonceMap[candidateRepBlock.Nonce]++//Change
					IDToNonce[shard.GlobalGroupMems[candidateRepBlock.ID].InShardId] = candidateRepBlock.Nonce
					if pow.Validate(candidateRepBlock.Nonce) {
						nonce = candidateRepBlock.Nonce
						hash = candidateRepBlock.Hash
						flag = false
						RepPowRxValidate <- candidateRepBlock
						fmt.Println("validate PoW from others - true")
						return nonce, hash, flag
					} else {
						fmt.Println("validate PoW from others - false")
					}
				}
			}
		default:
			{
				data := pow.prepareData(nonce)
				hash = sha256.Sum256(data)
				hashInt.SetBytes(hash[:])

				if hashInt.Cmp(pow.Target) == -1 {
					return nonce, hash, flag
				} else {
					nonce++
				}
			}
		}
	}

	return nonce, hash, flag
}

// Validate RepBlock's PoW
func (pow *ProofOfWork) Validate(nonce int) bool {
	var hashInt big.Int

	data := pow.prepareData(nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.Target) == -1

	return isValid
}
