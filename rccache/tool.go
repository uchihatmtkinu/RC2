package rccache

import (
	"fmt"

	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/shard"
)

//UTXOSet is the set of utxo in database
type UTXOSet struct {
	Cnt    uint32
	Data   []basic.OutType
	Stat   []uint32
	Remain uint32
	viewer uint32
}

//Encode is to serial
func (u *UTXOSet) Encode() []byte {
	var tmp []byte
	basic.EncodeInt(&tmp, u.Cnt)
	for i := uint32(0); i < u.Cnt; i++ {
		u.Data[i].OutToData(&tmp)
	}
	for i := uint32(0); i < u.Cnt; i++ {
		basic.EncodeInt(&tmp, u.Stat[i])
	}
	basic.EncodeInt(&tmp, u.Remain)
	return tmp
}

//Decode is to deserial
func (u *UTXOSet) Decode(tmp *[]byte) error {
	basic.DecodeInt(tmp, &u.Cnt)
	u.Data = make([]basic.OutType, u.Cnt)
	u.Stat = make([]uint32, u.Cnt)
	for i := uint32(0); i < u.Cnt; i++ {
		u.Data[i].DataToOut(tmp)
	}
	for i := uint32(0); i < u.Cnt; i++ {
		basic.DecodeInt(tmp, &u.Stat[i])
	}
	basic.DecodeInt(tmp, &u.Remain)
	if len(*tmp) != 0 {
		return fmt.Errorf("Decode utxoset error: Invalid length")
	}
	return nil
}

//ByteSlice returns a slice of a integer
func byteSlice(x uint32) []byte {
	var tmp []byte
	basic.EncodeInt(&tmp, x)
	return tmp
}

//GenerateTx is to generate a new tx
func GenerateTx(x int, y int, z uint32, seed int64, k uint32) *basic.Transaction {
	var tmp basic.Transaction
	tmp.New(0, seed, k)
	var b basic.OutType
	b.Address = shard.GlobalGroupMems[y].RealAccount.AddrReal
	b.Value = z
	var a basic.InType
	a.Init()
	a.PrevTx = shard.GlobalGroupMems[x].RealAccount.AddrReal
	a.Index = z
	tmp.AddOut(b)
	tmp.AddIn(a)
	tmp.Hash = tmp.HashTx()
	tmp.SignTx(0, &shard.GlobalGroupMems[x].RealAccount.Pri)
	return &tmp
}

//SendingChan is to send the signal on channel
func SendingChan(tmp *chan bool) {
	*tmp <- true
}
