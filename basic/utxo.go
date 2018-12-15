package basic

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/uchihatmtkinu/RC/base58"
	"github.com/uchihatmtkinu/RC/cryptonew"
	"github.com/uchihatmtkinu/RC/gVar"
)

//Init initial the big.Int parameters
func (a *InType) Init() {
	a.PukX = new(big.Int)
	a.PukY = new(big.Int)
}

//Acc returns whether the input address is an account or utxo
func (a *InType) Acc() bool {
	//Mark-UTXO
	//return cryptonew.Verify(a.Puk(), a.PrevTx)
	return true
}

//ShardIndex returns the target shard of the input address
func (a *InType) ShardIndex() uint32 {
	//tmp := cryptonew.GenerateAddr(a.Puk())
	return uint32(a.PrevTx[0]) % gVar.ShardCnt
}

//ShardIndex returns the shard index
func ShardIndex(str [32]byte) uint32 {
	return uint32(str[0]) % gVar.ShardCnt
}

//ShardIndex returns the target shard of the output address
func (a *OutType) ShardIndex() uint32 {
	return uint32(a.Address[0]) % gVar.ShardCnt
}

//Byte return the []byte of the input address used for hash
func (a *InType) Byte(b *[]byte) {
	EncodeByteL(b, a.PrevTx[:], 32)
	EncodeInt(b, a.Index)
}

//Puk returns the public key
func (a *InType) Puk() ecdsa.PublicKey {
	var tmp ecdsa.PublicKey
	tmp.Curve = elliptic.P256()
	tmp.X = a.PukX
	tmp.Y = a.PukY
	return tmp
}

//VerifyIn using the UTXO to verify the in address
func (a *InType) VerifyIn(b *OutType, h [32]byte) bool {
	if !cryptonew.Verify(a.Puk(), b.Address) {
		fmt.Println("UTXO.VerifyIn address doesn't match")
		return false
	}
	tmpPuk := a.Puk()
	if !a.Sig.Verify(h[:], &tmpPuk) {
		fmt.Println("UTXO.VerifyIn signature incorrect")
		return false
	}
	return true
}

//SignTxIn make the signature given the transaction
func (a *InType) SignTxIn(prk *ecdsa.PrivateKey, h [32]byte) {
	a.PukX.Set(prk.PublicKey.X)
	a.PukY.Set(prk.PublicKey.Y)
	a.Sig.Sign(h[:], prk)
}

//OutToData converts the output address data into bytes
func (a *OutType) OutToData(b *[]byte) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, a.Value)
	*b = append(*b, buf.Bytes()...)
	*b = append(*b, a.Address[:]...)
}

//DataToOut converts bytes into output address data
func (a *OutType) DataToOut(data *[]byte) error {
	err := Decode(data, &a.Value)
	if err != nil {
		return fmt.Errorf("Out Value read failed: %s", err)
	}
	err = Decode(data, &a.Address)
	if err != nil {
		return fmt.Errorf("Out Address read failed: %s", err)
	}
	return nil
}

//InToData converts the input address data into bytes
func (a *InType) InToData(b *[]byte) {
	Encode(b, &a.PrevTx)
	Encode(b, &a.Index)
	Encode(b, &a.Sig)
	Encode(b, a.PukX)
	Encode(b, a.PukY)
}

//DataToIn converts bytes into input address data
func (a *InType) DataToIn(data *[]byte) error {
	a.Init()
	err := Decode(data, &a.PrevTx)
	if err != nil {
		return fmt.Errorf("Input PrevHash read failed: %s", err)
	}
	err = Decode(data, &a.Index)
	if err != nil {
		return fmt.Errorf("Input Index read failed: %s", err)
	}
	err = Decode(data, &a.Sig)
	if err != nil {
		return fmt.Errorf("Input Signature read failed: %s", err)
	}
	err = Decode(data, a.PukX)
	if err != nil {
		return fmt.Errorf("Input Publick Key read failed: %s", err)
	}
	err = Decode(data, a.PukY)
	if err != nil {
		return fmt.Errorf("Input Publick Key read failed: %s", err)
	}
	return err
}

//Print is
func (a *InType) Print() {
	fmt.Println("InType: ", base58.Encode(a.PrevTx[:]), ", index: ", a.Index)
}

//Print is
func (a *OutType) Print() {
	fmt.Println("OutType: ", base58.Encode(a.Address[:]), ", index: ", a.Value)
}
