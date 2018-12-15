package basic

import (
	"crypto/ecdsa"
	"fmt"
	"math/rand"
	"time"

	"github.com/uchihatmtkinu/RC/base58"
)

//HashTx is come out the hash
func (a *Transaction) HashTx() [32]byte {
	var tmp []byte
	EncodeInt(&tmp, a.Timestamp)
	EncodeInt(&tmp, a.TxinCnt)
	EncodeInt(&tmp, a.TxoutCnt)
	EncodeInt(&tmp, a.Kind)
	EncodeInt(&tmp, a.Locktime)
	for i := uint32(0); i < a.TxinCnt; i++ {
		a.In[i].Byte(&tmp)
	}
	for i := uint32(0); i < a.TxoutCnt; i++ {
		a.Out[i].OutToData(&tmp)
	}
	tmpHash := new([32]byte)
	DoubleHash256(&tmp, tmpHash)
	return *tmpHash
}

//SignTx sign the ith in-address with the private key
func (a *Transaction) SignTx(i uint32, prk *ecdsa.PrivateKey) error {
	if a.TxinCnt <= i {
		return fmt.Errorf("Tx.SignTx in-address outrange")
	}
	a.In[i].SignTxIn(prk, a.Hash)
	return nil
}

//VerifyTx sign the ith in-address with the private key
func (a *Transaction) VerifyTx(i uint32, b *OutType) bool {
	if a.TxinCnt <= i {
		fmt.Println("Tx.VerifyTx in-address outrange")
		return false
	}

	return a.In[i].VerifyIn(b, a.Hash)
}

//Encode converts the transaction into bytes
func (a *Transaction) Encode(tmp *[]byte) {
	Encode(tmp, a.Timestamp)
	Encode(tmp, a.TxinCnt)
	Encode(tmp, a.TxoutCnt)
	Encode(tmp, a.Kind)
	Encode(tmp, a.Locktime)
	Encode(tmp, &a.Hash)
	for i := uint32(0); i < a.TxinCnt; i++ {
		a.In[i].InToData(tmp)
		//EncodeByte(&tmp, &xxx)
	}
	for i := uint32(0); i < a.TxoutCnt; i++ {
		a.Out[i].OutToData(tmp)
	}
}

//Decode decodes the packets into transaction format
func (a *Transaction) Decode(buf *[]byte) error {
	//buf := *data

	err := Decode(buf, &a.Timestamp)
	if err != nil {
		return fmt.Errorf("TX timestamp Read failed: %s", err)
	}
	err = Decode(buf, &a.TxinCnt)
	if err != nil {
		return fmt.Errorf("TX TxinCnt Read failed; %s", err)
	}
	err = Decode(buf, &a.TxoutCnt)
	if err != nil {
		return fmt.Errorf("TX TxoutCnt Read failed: %s", err)
	}
	err = Decode(buf, &a.Kind)
	if err != nil {
		return fmt.Errorf("TX Kind Read failed: %s", err)
	}
	err = Decode(buf, &a.Locktime)
	if err != nil {
		return fmt.Errorf("TX Locktime Read failed: %s", err)
	}
	err = Decode(buf, &a.Hash)
	if err != nil {
		return fmt.Errorf("TX hash Read failed: %s", err)
	}
	a.In = make([]InType, 0, a.TxinCnt)
	for i := uint32(0); i < a.TxinCnt; i++ {
		//var tmpArray *[]byte
		var tmpIn InType
		err = tmpIn.DataToIn(buf)
		if err != nil {
			return fmt.Errorf("Input Address Read failed-%d: %s", i, err)
		}
		a.In = append(a.In, tmpIn)
	}
	a.Out = make([]OutType, 0, a.TxoutCnt)
	for i := uint32(0); i < a.TxoutCnt; i++ {
		//var tmpArray *[]byte
		var tmpOut OutType
		err = tmpOut.DataToOut(buf)
		if err != nil {
			return fmt.Errorf("Output Address Read failed-%d: %s", i, err)
		}
		a.Out = append(a.Out, tmpOut)
	}
	return nil
}

//New is to initialize a transaction
func (a *Transaction) New(kind int, seed int64, writer uint32) error {
	rand.Seed(seed)
	a.Timestamp = uint64(time.Now().Unix()) + rand.Uint64()
	a.Locktime = writer * 10
	a.TxinCnt = 0
	a.TxoutCnt = 0
	a.Kind = uint32(kind)
	return nil
}

//MakeTx implements the method to create a new transaction
func MakeTx(a *[]InType, b *[]OutType, out *Transaction, kind int) error {
	if out == nil {
		return fmt.Errorf("Basic.MakeTx, null transaction")
	}
	out.Timestamp = uint64(time.Now().Unix())
	out.TxinCnt = uint32(len(*a))
	out.TxoutCnt = uint32(len(*b))
	out.Kind = uint32(kind)
	out.In = make([]InType, 0, out.TxinCnt)
	for i := 0; i < int(out.TxinCnt); i++ {
		out.In = append(out.In, (*a)[i])
	}
	out.Out = make([]OutType, 0, out.TxoutCnt)
	for i := 0; i < int(out.TxoutCnt); i++ {
		out.Out = append(out.Out, (*b)[i])
	}
	out.Hash = out.HashTx()
	return nil
}

//AddIn increases one input of transaction a
func (a *Transaction) AddIn(b InType) {
	a.TxinCnt++
	a.In = append(a.In, b)
}

//AddOut increases one output of transaction a
func (a *Transaction) AddOut(b OutType) {
	a.TxoutCnt++
	a.Out = append(a.Out, b)
}

//Print is
func (a *Transaction) Print() {
	fmt.Println("-----------Transaction-----------------")
	fmt.Println("Transaction: Time: ", a.Timestamp, " InCnt: ", a.TxinCnt, " OutCnt: ", a.TxoutCnt)
	fmt.Println("Kind: ", a.Kind)
	fmt.Println("Hash: ", base58.Encode(a.Hash[:]))
	for i := uint32(0); i < a.TxinCnt; i++ {
		fmt.Print(i, " ")
		a.In[i].Print()
	}
	for i := uint32(0); i < a.TxoutCnt; i++ {
		fmt.Print(i, " ")
		a.Out[i].Print()
	}
	fmt.Println("-----------Transaction end-----------------")
}

//New is to initialize a transactionbatch
func (a *TransactionBatch) New(b *[]Transaction) error {
	a.TxCnt = uint32(len(*b))
	a.TxArray = make([]Transaction, a.TxCnt)
	copy(a.TxArray, *b)
	return nil
}

//Add is to add a transactionbatch
func (a *TransactionBatch) Add(b *Transaction) error {
	a.TxCnt++
	a.TxArray = append(a.TxArray, *b)
	return nil
}

//Encode is to encode a transactionbatch
func (a *TransactionBatch) Encode() []byte {
	tmp := make([]byte, 0)
	Encode(&tmp, a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		a.TxArray[i].Encode(&tmp)
	}
	return tmp
}

//Decode is to encode a transactionbatch
func (a *TransactionBatch) Decode(buf *[]byte) error {
	err := Decode(buf, &a.TxCnt)
	if err != nil {
		return fmt.Errorf("TxBatch failed: %s", err)
	}
	a.TxArray = make([]Transaction, a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		err = a.TxArray[i].Decode(buf)
		if err != nil {
			return fmt.Errorf("TxBatch failed: %s", err)
		}
	}
	return nil
}
