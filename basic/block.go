package basic

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"

	"github.com/uchihatmtkinu/RC/base58"
	"github.com/uchihatmtkinu/RC/gVar"
)

//Hash generates the 32bits hash of one Tx block
func (a *TxBlock) Hash() [32]byte {
	tmp1 := make([]byte, 0, 136)
	tmp1 = append(byteSlice(a.ID), a.PrevHash[:]...)
	if a.Kind == 0 || a.Kind == 1 {
		tmp1 = append(tmp1, a.MerkleRoot[:]...)
	} else {
		for i := 0; i < len(a.TxHash); i++ {
			tmp1 = append(tmp1, a.TxHash[i][:]...)
		}
	}
	EncodeInt(&tmp1, a.Timestamp)
	var b [32]byte
	DoubleHash256(&tmp1, &b)
	return b
}

//GenMerkTree generates the merkleroot tree given the transactions
func GenMerkTree(d *[]Transaction, out *[32]byte) error {
	if len(*d) == 0 {
		return nil
	}
	if len(*d) == 1 {
		tmp := (*d)[0].Hash[:]
		DoubleHash256(&tmp, out)
	} else {
		l := len(*d)
		d1 := (*d)[:l/2]
		d2 := (*d)[l/2:]
		var out1, out2 [32]byte
		GenMerkTree(&d1, &out1)
		GenMerkTree(&d2, &out2)
		tmp := append(out1[:], out2[:]...)
		DoubleHash256(&tmp, out)
	}
	return nil
}

//Verify verify the signature of the Txblock
func (a *TxBlock) Verify(puk *ecdsa.PublicKey) (bool, error) {

	tmp := a.Hash()
	//Verify the hash, the cnt of in and out address
	if tmp != a.HashID {
		return false, fmt.Errorf("VerifyTxBlock Invalid parameter")
	}
	if !a.Sig.Verify(a.HashID[:], puk) {
		return false, fmt.Errorf("VerifyTxBlock Invalid signature")
	}
	return true, nil
}

//NewGensisTxBlock is the gensis block
func NewGensisTxBlock() TxBlock {
	var a TxBlock
	a.ID = 0
	a.TxCnt = 0
	a.HashID = sha256.Sum256([]byte(GenesisTxBlock))
	a.Height = 0
	return a
}

//NewGensisFinalTxBlock is the gensis block
func NewGensisFinalTxBlock(shardID uint32) TxBlock {
	var a TxBlock
	a.ID = 0
	a.TxCnt = 0
	a.ShardID = shardID
	a.Kind = 1
	a.HashID = sha256.Sum256(append([]byte(GenesisTxBlock), []byte(strconv.Itoa(int(shardID)))...))
	a.Height = 0
	return a
}

//MakeTxBlock creates the transaction blocks given verified transactions
func (a *TxBlock) MakeTxBlock(ID uint32, b *[]Transaction, preHash [32]byte, prk *ecdsa.PrivateKey, h uint32, kind uint32, preFH *[32]byte, shardID uint32) error {
	a.ID = ID
	a.Kind = kind
	a.PrevHash = preHash
	a.Timestamp = time.Now().Unix()
	a.TxCnt = uint32(len(*b))
	a.Height = h
	a.TxArray = *b
	a.TxArrayX = make([][SHash]byte, a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		a.TxArrayX[i] = HashCut(a.TxArray[i].Hash)
	}
	GenMerkTree(&a.TxArray, &a.MerkleRoot)
	if kind == 1 {
		a.PrevFinalHash = *preFH
		a.ShardID = shardID
	}
	a.HashID = a.Hash()
	a.Sig.Sign(a.HashID[:], prk)
	return nil
}

//MakeStartBlock creates the transaction blocks given verified transactions
func (a *TxBlock) MakeStartBlock(ID uint32, b *[][32]byte, preHash [32]byte, prk *ecdsa.PrivateKey, h uint32) error {
	a.ID = ID
	a.Kind = 2
	a.PrevHash = preHash
	a.Timestamp = time.Now().Unix()
	a.TxCnt = 0
	a.Height = h
	a.TxHash = *b
	a.HashID = a.Hash()
	a.Sig.Sign(a.HashID[:], prk)
	return nil
}

//Serial outputs a serial of []byte
func (a *TxBlock) Serial() []byte {
	var tmp []byte
	a.Encode(&tmp, 1)
	return tmp
}

//Transform is to minimize the txdecset size
func (a *TxBlock) Transform() error {
	a.TxArrayX = make([][SHash]byte, a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		a.TxArrayX[i] = HashCut(a.TxArray[i].Hash)
	}
	return nil
}

//Encode converts the block data into bytes
func (a *TxBlock) Encode(tmp *[]byte, full int) {
	Encode(tmp, a.ID)
	Encode(tmp, a.Level)
	Encode(tmp, a.Sender)
	Encode(tmp, &a.PrevHash)
	Encode(tmp, &a.HashID)
	Encode(tmp, &a.MerkleRoot)
	Encode(tmp, a.Timestamp)
	Encode(tmp, a.Height)
	Encode(tmp, a.TxCnt)
	Encode(tmp, a.Kind)
	if a.Kind == 1 {
		Encode(tmp, &a.PrevFinalHash)
		Encode(tmp, &a.ShardID)
	}
	if a.Kind == 2 {
		for i := uint32(0); i < gVar.ShardCnt; i++ {
			Encode(tmp, &a.TxHash[i])
		}
	}
	if full == 1 {
		for i := uint32(0); i < a.TxCnt; i++ {
			a.TxArray[i].Encode(tmp)
		}
	} else {
		for i := uint32(0); i < a.TxCnt; i++ {
			Encode(tmp, &a.TxArrayX[i])
		}
	}
	if a.Kind != 1 {
		if a.HashID != sha256.Sum256([]byte(GenesisTxBlock)) {
			Encode(tmp, &a.Sig)
		}
	} else {
		if a.HashID != sha256.Sum256(append([]byte(GenesisTxBlock), []byte(strconv.Itoa(int(a.ShardID)))...)) {
			Encode(tmp, &a.Sig)
		}
	}
}

//Decode converts bytes into block data
func (a *TxBlock) Decode(buf *[]byte, full int) error {
	err := Decode(buf, &a.ID)
	if err != nil {
		return fmt.Errorf("TxBlock ID failed %s", err)
	}
	err = Decode(buf, &a.Level)
	if err != nil {
		return fmt.Errorf("TxBlock Level failed %s", err)
	}
	err = Decode(buf, &a.Sender)
	if err != nil {
		return fmt.Errorf("TxBlock Sender failed %s", err)
	}
	err = Decode(buf, &a.PrevHash)
	if err != nil {
		return fmt.Errorf("TxBlock PrevHash failed %s", err)
	}
	err = Decode(buf, &a.HashID)
	if err != nil {
		return fmt.Errorf("TxBlock HashID failed %s", err)
	}
	err = Decode(buf, &a.MerkleRoot)
	if err != nil {
		return fmt.Errorf("TxBlock MerkleRoot failed: %s", err)
	}
	err = Decode(buf, &a.Timestamp)
	if err != nil {
		return fmt.Errorf("TxBlock Timestamp failed: %s", err)
	}
	err = Decode(buf, &a.Height)
	if err != nil {
		return fmt.Errorf("TxBlock Height failed: %s", err)
	}
	err = Decode(buf, &a.TxCnt)
	if err != nil {
		return fmt.Errorf("TxBlock TxCnt failed: %s", err)
	}
	err = Decode(buf, &a.Kind)
	if err != nil {
		return fmt.Errorf("TxBlock Kind failed: %s", err)
	}
	if a.Kind == 1 {
		err = Decode(buf, &a.PrevFinalHash)
		if err != nil {
			return fmt.Errorf("TxBlock PrevFinalHash failed: %s", err)
		}
		err = Decode(buf, &a.ShardID)
		if err != nil {
			return fmt.Errorf("TxBlock ShardID failed: %s", err)
		}
	}
	if a.Kind == 2 {
		a.TxHash = make([][32]byte, gVar.ShardCnt)
		for i := uint32(0); i < gVar.ShardCnt; i++ {
			err = Decode(buf, &a.TxHash[i])
			if err != nil {
				return fmt.Errorf("TxBlock TxHash failed: %s", err)
			}
		}
	}
	if full == 1 {
		a.TxArray = make([]Transaction, a.TxCnt)
		for i := uint32(0); i < a.TxCnt; i++ {
			err = a.TxArray[i].Decode(buf)
			if err != nil {
				return fmt.Errorf("TxBlock decode Tx failed-%d: %s", i, err)
			}
		}
	} else {
		a.TxArrayX = make([][SHash]byte, a.TxCnt)
		for i := uint32(0); i < a.TxCnt; i++ {
			err = Decode(buf, &a.TxArrayX[i])
			if err != nil {
				return fmt.Errorf("TxBlock decode Tx failed-%d: %s", i, err)
			}
		}
	}
	if a.Kind == 0 || a.Kind == 3 {
		if a.HashID != sha256.Sum256([]byte(GenesisTxBlock)) {
			err = Decode(buf, &a.Sig)
			if err != nil {
				return fmt.Errorf("TxBlock Signature failed: %s", err)
			}
		}
	} else {
		if a.HashID != sha256.Sum256(append([]byte(GenesisTxBlock), []byte(strconv.Itoa(int(a.ShardID)))...)) {
			err = Decode(buf, &a.Sig)
			if err != nil {
				return fmt.Errorf("TxBlock Signature failed: %s", err)
			}
		}
	}
	/*if len(*buf) != 0 {
		return fmt.Errorf("TxBlock decode failed: With extra bits")
	}*/
	return nil
}

//Print output the block data
func (a *TxBlock) Print() {
	fmt.Println("----------TxBlock-----------")
	fmt.Println("ID:", a.ID, "Hash:", base58.Encode(a.HashID[:]), "PrevH:", base58.Encode(a.PrevHash[:]))
	fmt.Println("Height:", a.Height, "TxCnt:", a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		a.TxArray[i].Print()
	}
}
