package basic

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"

	"github.com/uchihatmtkinu/RC/base58"
)

//Hash returns the ID of the TxList
func (a *TxList) Hash() [32]byte {
	tmp := make([]byte, 0, 4+a.TxCnt*32)
	tmp = byteSlice(a.Round)
	for i := uint32(0); i < a.TxCnt; i++ {
		tmp = append(tmp, a.TxArray[i][:]...)
	}
	return sha256.Sum256(tmp)
}

//Sign signs the TxList with the leader's private key
func (a *TxList) Sign(prk *ecdsa.PrivateKey) {
	a.HashID = a.Hash()
	a.Sig.Sign(a.HashID[:], prk)
}

//Verify verify the signature
func (a *TxList) Verify(puk *ecdsa.PublicKey) bool {
	if a.Hash() != a.HashID {
		return false
	}
	return a.Sig.Verify(a.HashID[:], puk)
}

//Set init an instance of TxList given those parameters
func (a *TxList) Set(ID uint32, height uint32) {
	a.ID = ID
	a.TxCnt = 0
	a.Round = height
	a.TxArray = nil
}

//AddTx adds the tx into transaction list
func (a *TxList) AddTx(tx *Transaction) {
	a.TxCnt++
	a.TxArray = append(a.TxArray, tx.Hash)
	a.TxArrayX = append(a.TxArrayX, HashCut(tx.Hash))
}

//Encode returns the byte of a TxList
func (a *TxList) Encode(tmp *[]byte) {
	Encode(tmp, a.ID)
	Encode(tmp, &a.HashID)
	Encode(tmp, a.Round)
	Encode(tmp, a.Level)
	Encode(tmp, a.Sender)
	Encode(tmp, a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		Encode(tmp, a.TxArray[i])
	}
	Encode(tmp, &a.Sig)
}

//Decode decodes the TxList with []byte
func (a *TxList) Decode(buf *[]byte) error {
	err := Decode(buf, &a.ID)
	if err != nil {
		return fmt.Errorf("TxList ID decode failed: %s", err)
	}
	err = Decode(buf, &a.HashID)
	if err != nil {
		return fmt.Errorf("TxList HashID decode failed: %s", err)
	}
	err = Decode(buf, &a.Round)
	if err != nil {
		return fmt.Errorf("TxList Round decode failed: %s", err)
	}
	err = Decode(buf, &a.Level)
	if err != nil {
		return fmt.Errorf("TxList Level decode failed: %s", err)
	}
	err = Decode(buf, &a.Sender)
	if err != nil {
		return fmt.Errorf("TxList Sender decode failed: %s", err)
	}
	err = Decode(buf, &a.TxCnt)
	if err != nil {
		return fmt.Errorf("TxList TxCnt decode failed: %s", err)
	}
	a.TxArray = make([][32]byte, a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		err = Decode(buf, &a.TxArray[i])
		if err != nil {
			return fmt.Errorf("TxList Tx decode failed-%d: %s", i, err)
		}
	}
	err = Decode(buf, &a.Sig)
	if err != nil {
		return fmt.Errorf("TxList signature decode failed: %s", err)
	}
	if len(*buf) != 0 {
		return fmt.Errorf("TxList decode failed: With extra bits")
	}
	return nil
}

//Print is
func (a *TxList) Print() {
	fmt.Println("---------TxList---------------------")
	fmt.Println("TxList: ID: ", a.ID, " TxCnt: ", a.TxCnt)
	fmt.Println("Hash: ", base58.Encode(a.HashID[:]))
	for i := uint32(0); i < a.TxCnt; i++ {
		//fmt.Println(i, ":", base58.Encode(a.TxArray[i][:]))
	}
	fmt.Println("---------TxList----END---------------")
}
