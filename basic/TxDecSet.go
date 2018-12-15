package basic

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"

	"github.com/uchihatmtkinu/RC/gVar"
)

//Hash returns the ID of the TxList
func (a *TxDecSet) Hash() [32]byte {
	tmp := make([]byte, 0, a.TxCnt*32)
	for i := uint32(0); i < a.TxCnt; i++ {
		tmp = append(tmp, a.TxArray[i][:]...)
	}
	return sha256.Sum256(tmp)
}

//Sign signs the TxDecSet
func (a *TxDecSet) Sign(prk *ecdsa.PrivateKey) {
	tmp := make([]byte, 0, 36+int(a.TxCnt/4*a.MemCnt))
	tmp = append(byteSlice(a.ID), a.HashID[:]...)
	for i := uint32(0); i < a.MemCnt; i++ {
		tmp = append(tmp, a.MemD[i].Decision...)
	}
	tmpHash := sha256.Sum256(tmp)
	a.Sig.Sign(tmpHash[:], prk)
}

//Verify verifies the TxDecSet
func (a *TxDecSet) Verify(puk *ecdsa.PublicKey) bool {
	tmp := make([]byte, 0, a.TxCnt*32)
	for i := uint32(0); i < a.TxCnt; i++ {
		tmp = append(tmp, a.TxArray[i][:]...)
	}
	if sha256.Sum256(tmp) != a.HashID {
		return false
	}
	tmp = make([]byte, 0, 36+int(a.TxCnt/4*a.MemCnt))
	tmp = append(byteSlice(a.ID), a.HashID[:]...)
	for i := uint32(0); i < a.MemCnt; i++ {
		tmp = append(tmp, a.MemD[i].Decision...)
	}
	tmpHash := sha256.Sum256(tmp)
	return a.Sig.Verify(tmpHash[:], puk)
}

//Set init an instance of TxDecSet given those parameters
func (a *TxDecSet) Set(b *TxList, x uint32, y int) {
	a.ID = b.ID
	a.HashID = b.HashID
	a.MemCnt = 0
	a.Round = b.Round
	a.ShardIndex = x
	if y == 1 {
		a.TxCnt = 0
		a.TxArray = nil
	} else {
		a.TxCnt = b.TxCnt
		a.TxArray = make([][32]byte, 0, a.TxCnt)
		a.TxArrayX = make([][SHash]byte, 0, a.TxCnt)
		for i := uint32(0); i < a.TxCnt; i++ {
			a.TxArray = append(a.TxArray, b.TxArray[i])
			a.TxArrayX = append(a.TxArrayX, HashCut(b.TxArray[i]))
		}
	}
}

//Add adds a TxDecision
func (a *TxDecSet) Add(b *TxDecision) {
	a.MemCnt++
	a.MemD = append(a.MemD, *b)
}

//ResultMiner is the specific result of a miner
func (a *TxDecSet) ResultMiner(index uint32, miner uint32) byte {
	x := index / 8
	y := byte(index % 8)
	return byte((a.MemD[miner].Decision[x] >> y) & 1)
}

//Result is the result of the index-th transaction
func (a *TxDecSet) Result(index uint32) bool {
	x := index / 8
	y := byte(index % 8)
	var ans uint32
	for i := uint32(0); i < a.MemCnt; i++ {
		ans = ans + uint32((a.MemD[i].Decision[x]>>y)&1)
	}
	if ans > (gVar.ShardSize-1)/2 {
		return true
	}
	return true
	//return false
}

//Encode encode the TxDecSet into []byte
func (a *TxDecSet) Encode(tmp *[]byte) {
	Encode(tmp, a.ID)
	Encode(tmp, a.Round)
	Encode(tmp, a.Level)
	Encode(tmp, a.Sender)
	Encode(tmp, &a.HashID)
	Encode(tmp, a.MemCnt)
	Encode(tmp, a.TxCnt)
	Encode(tmp, a.ShardIndex)
	for i := uint32(0); i < a.MemCnt; i++ {
		a.MemD[i].Encode(tmp)
	}
	for i := uint32(0); i < a.TxCnt; i++ {
		Encode(tmp, &a.TxArray[i])
	}
	Encode(tmp, &a.Sig)
}

//Decode decode the []byte into TxDecSet
func (a *TxDecSet) Decode(buf *[]byte) error {
	err := Decode(buf, &a.ID)
	if err != nil {
		return fmt.Errorf("TxDecSet ID decode failed: %s", err)
	}
	err = Decode(buf, &a.Round)
	if err != nil {
		return fmt.Errorf("TxDecSet Round decode failed: %s", err)
	}
	err = Decode(buf, &a.Level)
	if err != nil {
		return fmt.Errorf("TxDecSet Level decode failed: %s", err)
	}
	err = Decode(buf, &a.Sender)
	if err != nil {
		return fmt.Errorf("TxDecSet Sender decode failed: %s", err)
	}
	err = Decode(buf, &a.HashID)
	if err != nil {
		return fmt.Errorf("TxDecSet HashID decode failed: %s", err)
	}
	err = Decode(buf, &a.MemCnt)
	if err != nil {
		return fmt.Errorf("TxDecSet MemCnt decode failed: %s", err)
	}
	err = Decode(buf, &a.TxCnt)
	if err != nil {
		return fmt.Errorf("TxDecSet TxCnt decode failed: %s", err)
	}
	err = Decode(buf, &a.ShardIndex)
	if err != nil {
		return fmt.Errorf("TxDecSet ShardIndex decode failed: %s", err)
	}
	a.MemD = make([]TxDecision, a.MemCnt)
	for i := uint32(0); i < a.MemCnt; i++ {
		err = a.MemD[i].Decode(buf)
		if err != nil {
			return fmt.Errorf("TxDecSet MemDecision decode failed-%d: %s", i, err)
		}
	}
	a.TxArray = make([][32]byte, a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		err = Decode(buf, &a.TxArray[i])
		if err != nil {
			return fmt.Errorf("TxDecSet TxArray decode failed-%d: %s", i, err)
		}
	}
	err = Decode(buf, &a.Sig)
	if err != nil {
		return fmt.Errorf("TxDecSet Signature decode failed: %s", err)
	}
	if len(*buf) != 0 {
		return fmt.Errorf("TxDecSet decode failed: With extra bits")
	}
	return nil
}

//Print is
func (a *TxDecSet) Print() {
	fmt.Println("TxDecision: ID: ", a.ID, " TxCnt: ", a.TxCnt, " MemCnt: ", a.MemCnt)
	fmt.Println()
}
