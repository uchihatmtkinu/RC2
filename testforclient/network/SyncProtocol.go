package network

import (
	"bytes"
	"encoding/gob"
	"log"
	"sync"
	"time"

	"math/rand"

	"fmt"

	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/base58"
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

//sbrxCounter count for the block receive
//var sbrxCounter safeCounter

//aski ask user i for sync, -1 means done, otherwise ask i+1
var aski []int

//SyncProcess processing sync after one epoch
func SyncProcess(ms *[]shard.MemShard) {

	CurrentEpoch++
	fmt.Println("Sync Began")
	startDone = false
	//waitgroup for all goroutines done
	var wg sync.WaitGroup
	aski = make([]int, int(gVar.ShardCnt))
	rand.Seed(int64(shard.MyMenShard.Shard*3000+shard.MyMenShard.InShardId) + time.Now().UTC().UnixNano())
	//intilizeMaskBit(&syncmask, int((gVar.ShardCnt+7)>>3),false)
	for i := 0; i < int(gVar.ShardCnt); i++ {
		syncSBCh[i] = make(chan syncSBInfo, 10)
		syncTBCh[i] = make(chan syncTBInfo, 10)
		syncNotReadyCh[i] = make(chan bool, 10)
		aski[i] = rand.Intn(int(gVar.ShardSize))
	}
	SyncFlag = true

	for i := 0; i < int(gVar.ShardCnt); i++ {
		//if !maskBit(i, &syncmask) {
		if i != shard.MyMenShard.Shard {
			wg.Add(1)
			go SendSyncMessage((*ms)[shard.ShardToGlobal[i][aski[i]]].Address, "requestSync", syncRequestInfo{ID: MyGlobalID, Epoch: CurrentEpoch})
			go ReceiveSyncProcess(i, &wg, ms)
		}
		//}
	}
	wg.Wait()
	SyncFlag = false
	for i := 0; i < int(gVar.ShardCnt); i++ {
		close(syncSBCh[i])
		close(syncTBCh[i])
		close(syncNotReadyCh[i])
	}

	fmt.Println("Sync Finished")
	//ShardProcess()
}

//ReceiveSyncProcess listen to the block from shard k
func ReceiveSyncProcess(k int, wg *sync.WaitGroup, ms *[]shard.MemShard) {
	fmt.Println("wait for shard", k, " from user", shard.ShardToGlobal[k][aski[k]])
	defer wg.Done()

	//syncblock flag
	sbrxflag := true
	//txblock flag
	//TODO test
	tbrxflag := true
	//txBlock Transaction block
	//TODO test
	var txBlockMessage syncTBInfo
	//syncBlock SyncBlock
	var syncBlockMessage syncSBInfo
	for sbrxflag || tbrxflag {
		select {
		case syncBlockMessage = <-syncSBCh[k]:
			{
				if syncBlockMessage.Block.VerifyCoSignature(ms) {
					sbrxflag = false
					fmt.Println("Get cosi from", k)
				} else {
					//aski[k] = (aski[k] + 1) % int(gVar.ShardSize)
					fmt.Println("Verifyied cosi falied")
					sbrxflag = false
					//TODO test
					//tbrxflag = true
					//go SendSyncMessage((*ms)[shard.ShardToGlobal[k][aski[k]]].Address, "requestSync", syncRequestInfo{MyGlobalID, CurrentEpoch})
				}
			}
		//TODO test
		case txBlockMessage = <-syncTBCh[k]:
			tmpIndex := 0
			if gVar.ExperimentBadLevel != 0 {
				for shard.ShardToGlobal[k][tmpIndex] < int(gVar.ShardCnt*gVar.ShardSize/3) {
					tmpIndex++
				}
			}
			ok, err := txBlockMessage.Block.Verify(&(*ms)[shard.ShardToGlobal[k][tmpIndex]].RealAccount.Puk)
			if ok {
				tbrxflag = false
				tmpBlock := txBlockMessage.Block
				CacheDbRef.FB[k] = &tmpBlock
				fmt.Println("Get FB from", k)
			} else {
				fmt.Println("FinalTxBlock verify failed: ", err)
			}
		case <-syncNotReadyCh[k]:
			fmt.Println(time.Now(), "sleep for not ready")
			time.Sleep(timeSyncNotReadySleep)
			go SendSyncMessage((*ms)[shard.ShardToGlobal[k][aski[k]]].Address, "requestSync", syncRequestInfo{ID: MyGlobalID, Epoch: CurrentEpoch})
		case <-time.After(timeoutSync):
			{
				aski[k] = (aski[k] + 1) % int(gVar.ShardSize)
				fmt.Println(time.Now(), "wait for shard", k, " from user", shard.ShardToGlobal[k][aski[k]])
				go SendSyncMessage((*ms)[shard.ShardToGlobal[k][aski[k]]].Address, "requestSync", syncRequestInfo{ID: MyGlobalID, Epoch: CurrentEpoch})
			}
		}
	}
	if !sbrxflag && !tbrxflag {
		//add transaction block
		//TODO test
		CacheDbRef.Mu.Lock()
		CacheDbRef.GetFinalTxBlock(&txBlockMessage.Block)
		CacheDbRef.Mu.Unlock()
		//update reputation of members
		syncBlockMessage.Block.UpdateTotalRepInMS(ms)
		//add sync Block
		Reputation.MyRepBlockChain.AddSyncBlockFromOtherShards(&syncBlockMessage.Block, k)
	}
	fmt.Println("received sync from shard:", k)
}

//SendSyncMessage send cosi message
func SendSyncMessage(addr string, command string, message interface{}) {
	payload := gobEncode(message)
	request := append(commandToBytes(command), payload...)
	sendData(addr, request)
}

//HandleSyncNotReady rx challenge
func HandleSyncNotReady(request []byte) {
	var buff bytes.Buffer
	var payload syncNotReadyInfo
	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	syncNotReadyCh[shard.GlobalGroupMems[payload.ID].Shard] <- true
}

//HandleRequestSync handles the sync request
func HandleRequestSync(request []byte) {
	var buff bytes.Buffer
	var payload syncRequestInfo

	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	addr := shard.GlobalGroupMems[payload.ID].Address
	fmt.Println("Recived request sync from ", payload.ID)
	Reputation.CurrentSyncBlock.Mu.RLock()
	defer Reputation.CurrentSyncBlock.Mu.RUnlock()
	if payload.Epoch > Reputation.CurrentSyncBlock.Epoch {
		SendSyncMessage(addr, "syncNReady", syncNotReadyInfo{MyGlobalID, Reputation.CurrentSyncBlock.Epoch})
		return
	}
	if payload.Epoch == Reputation.CurrentSyncBlock.Epoch {
		//TODO test
		tmp := syncTBInfo{MyGlobalID, *(CacheDbRef.FB[CacheDbRef.HistoryShard[payload.Epoch]])}
		sendTxMessage(addr, "syncTB", tmp.Encode())
		SendSyncMessage(addr, "syncSB", syncSBInfo{MyGlobalID, *Reputation.CurrentSyncBlock.Block})
		return
	}

}

//HandleSyncSBMessage rx challenge
func HandleSyncSBMessage(request []byte) {
	var buff bytes.Buffer
	var payload syncSBInfo

	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	fmt.Println("SyncSBdata:", payload.ID, "Hash:", base58.Encode(payload.Block.Hash[:]))
	if err != nil {
		log.Panic(err)
	}
	syncSBCh[shard.GlobalGroupMems[payload.ID].Shard] <- payload
}

//HandleSyncTBMessage decodes the txblock
func HandleSyncTBMessage(request []byte) {
	var payload syncTBInfo
	tmpdata := make([]byte, len(request))
	copy(tmpdata, request)
	err := payload.Decode(&tmpdata)

	fmt.Println("SyncTBdata:", payload.ID, "Hash:", base58.Encode(payload.Block.HashID[:]))
	//payload.Block.Print()
	if err != nil {
		log.Panic(err)
	}
	syncTBCh[shard.GlobalGroupMems[payload.ID].Shard] <- payload

}

//Encode is the encode
func (s *syncTBInfo) Encode() []byte {
	var tmp []byte
	basic.Encode(&tmp, uint32(s.ID))
	tmp = append(tmp, s.Block.Serial()...)
	fmt.Println("syncTBInfo length: ", len(tmp))
	return tmp
}

//Decode is the decode
func (s *syncTBInfo) Decode(buf *[]byte) error {
	var tmp uint32
	basic.Decode(buf, &tmp)
	s.ID = int(tmp)
	err := s.Block.Decode(buf, 1)
	return err
}
