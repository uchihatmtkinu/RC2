package network

import (
	"fmt"

	"github.com/uchihatmtkinu/RC/gVar"

	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/shard"
)

//HandleRequestTxMM is handle the request of TxMM
func HandleRequestTxMM(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	var tmp txDecRev
	err := tmp.Decode(&data1)
	if err != nil {
		fmt.Println("RequestTxMM decoding error")
		return err
	}
	if tmp.Round >= CacheDbRef.PrevHeight && tmp.Round < CacheDbRef.PrevHeight+gVar.NumTxListPerEpoch {
		if BatchCache[tmp.Round-CacheDbRef.PrevHeight] != nil {
			xx := shard.GlobalGroupMems[tmp.ID].Shard
			go sendTxMessage(shard.GlobalGroupMems[tmp.ID].Address, "TxMM", BatchCache[tmp.Round-CacheDbRef.PrevHeight][xx].Encode())
		}
	}
	return nil
}

//HandleTxMMRec is handle request
func HandleTxMMRec(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	var tmp txDecRev
	err := tmp.Decode(&data1)
	if err != nil {
		fmt.Println("TxMMRec decoding error")
		return err
	}
	xx := tmp.Round
	if xx < CacheDbRef.PrevHeight || xx >= CacheDbRef.PrevHeight+gVar.NumTxListPerEpoch {
		return fmt.Errorf("TxMMRec epoch not match")
	}
	txMCh[xx-CacheDbRef.PrevHeight] <- tmp
	return nil
}

//Encode is encode
func (a *TxBatchInfo) Encode() []byte {
	var tmp []byte
	basic.Encode(&tmp, a.ID)
	basic.Encode(&tmp, a.ShardID)
	basic.Encode(&tmp, a.Epoch)
	basic.Encode(&tmp, a.Round)
	basic.Encode(&tmp, &a.Data)
	return tmp
}

//Decode is decode
func (a *TxBatchInfo) Decode(data *[]byte) error {
	data1 := make([]byte, len(*data))
	copy(data1, *data)
	err := basic.Decode(&data1, &a.ID)
	if err != nil {
		return err
	}
	err = basic.Decode(&data1, &a.ShardID)
	if err != nil {
		return err
	}
	err = basic.Decode(&data1, &a.Epoch)
	if err != nil {
		return err
	}
	err = basic.Decode(&data1, &a.Round)
	if err != nil {
		return err
	}
	err = basic.Decode(&data1, &a.Data)
	if err != nil {
		return err
	}
	return nil
}
