package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/Reputation/cosi"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

var readymask []byte
var SentLeaderReadyFlag = true

//ShardProcess is the process of sharding
func ShardProcess() {
	var beginShard shard.Instance

	Reputation.CurrentRepBlock.Mu.Lock()
	Reputation.CurrentRepBlock.Round = -1
	CurrentRepRound = -1
	Reputation.CurrentRepBlock.Mu.Unlock()

	shard.StartFlag = true
	shard.ShardToGlobal = make([][]int, gVar.ShardCnt)
	if CurrentEpoch != -1 {
		for i := 0; i < int(gVar.ShardSize*gVar.ShardCnt); i++ {
			shard.GlobalGroupMems[i].ClearRep()
		}
	}
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		shard.ShardToGlobal[i] = make([]int, gVar.ShardSize)

	}
	beginShard.GenerateSeed(&shard.PreviousSyncBlockHash)
	beginShard.Sharding(&shard.GlobalGroupMems, &shard.ShardToGlobal)
	//shard.MyMenShard = &shard.GlobalGroupMems[MyGlobalID]
	fmt.Println(time.Now(), CacheDbRef.ID, "Shard Calculated")
	LeaderAddr = shard.GlobalGroupMems[shard.ShardToGlobal[shard.MyMenShard.Shard][0]].Address
	CacheDbRef.Mu.Lock()
	if CurrentEpoch != -1 {
		CacheDbRef.PrevHeight = CacheDbRef.PrevHeight + gVar.NumTxListPerEpoch + 3
	}
	CacheDbRef.DB.ClearTx()
	CacheDbRef.ShardNum = uint32(shard.MyMenShard.Shard)
	CacheDbRef.Leader = uint32(shard.ShardToGlobal[shard.MyMenShard.Shard][0])
	CacheDbRef.HistoryShard = append(CacheDbRef.HistoryShard, CacheDbRef.ShardNum)
	if CurrentEpoch != -1 {
		CacheDbRef.Clear()

	}
	for i := 0; i < gVar.NumTxListPerEpoch; i++ {
		for j := uint32(0); j < gVar.ShardSize; j++ {
			CacheDbRef.RepCache[i][j] = 0
		}
	}
	CacheDbRef.Mu.Unlock()
	for i := uint32(0); i < gVar.NumTxListPerEpoch; i++ {
		BatchCache[i] = nil
	}

	//StopGetTx = make(chan bool, 1)
	close(Reputation.RepPowRxCh)
	Reputation.RepPowRxCh = make(chan Reputation.RepPowInfo, bufferSize)
	if shard.MyMenShard.Role == 1 {
		MinerReadyProcess()
	} else {
		LeaderReadyProcess(&shard.GlobalGroupMems)
		if CurrentEpoch != -1 {
			//go SendStartBlock(&shard.GlobalGroupMems)
		}
	}
	fmt.Println("shard finished")
	if CacheDbRef.ID == 0 {
		tmpStr := fmt.Sprint("Epoch", CurrentEpoch, ":")
		for i := uint32(0); i < gVar.ShardCnt*gVar.ShardSize; i++ {
			tmpStr = tmpStr + fmt.Sprint(shard.GlobalGroupMems[i].CalTotalRep(), " ")
		}
		sendTxMessage(gVar.MyAddress, "LogInfo", []byte(tmpStr))
	}
	if CurrentEpoch != -1 {
		FinalTxReadyCh <- true
	}

}

//LeaderReadyProcess leader use this
func LeaderReadyProcess(ms *[]shard.MemShard) {
	var readyMessage readyInfo
	var it *shard.MemShard
	var membermask []byte
	var leadermask []byte
	intilizeMaskBit(&membermask, (int(gVar.ShardSize)+7)>>3, cosi.Disabled)
	intilizeMaskBit(&leadermask, (int(gVar.ShardCnt)+7)>>3, cosi.Disabled)
	readyMember := 1
	readyLeader := 1
	//sent announcement
	for i := 1; i < int(gVar.ShardSize); i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		SendShardReadyMessage(it.Address, "readyAnnoun", readyInfo{MyGlobalID, CurrentEpoch})
	}
	//fmt.Println("wait for ready")
	//TODO modify int(gVar.ShardSize)/2
	cnt := 0
	timeoutflag := true
	SentLeaderReadyFlag = false
	for readyMember < int(gVar.ShardSize) && timeoutflag {
		select {
		case readyMessage = <-readyMemberCh:
			if readyMessage.Epoch == CurrentEpoch {
				readyMember++
				setMaskBit((*ms)[readyMessage.ID].InShardId, cosi.Enabled, &membermask)
				//fmt.Println("ReadyCount: ", readyCount)
			}
		case <-time.After(timeoutSync * 2):
			//fmt.Println("Wait shard signal time out")
			for i := 1; i < int(gVar.ShardSize); i++ {
				if maskBit(i, &membermask) == cosi.Disabled {
					it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
					fmt.Println("Resend shard ready to Member: ", shard.ShardToGlobal[shard.MyMenShard.Shard][i])
					SendShardReadyMessage(it.Address, "readyAnnoun", readyInfo{MyGlobalID, CurrentEpoch})
				}

			}
			cnt++
			if cnt > 5 && readyMember >= int(gVar.ShardSize*2/3) {
				timeoutflag = false
				fmt.Println("Timeout! Ready Member: ", readyMember, "/", gVar.ShardSize)
			}
		}
	}
	fmt.Println(time.Now(), "Shard is ready, sent to other shards")
	for i := 0; i < int(gVar.ShardCnt); i++ {
		if i != shard.MyMenShard.Shard {
			it = &(*ms)[shard.ShardToGlobal[i][0]]
			SendShardReadyMessage(it.Address, "leaderReady", readyInfo{shard.MyMenShard.Shard, CurrentEpoch})
		}
	}
	SentLeaderReadyFlag = true
	for readyLeader < int(gVar.ShardCnt) && timeoutflag {
		select {
		case readyMessage = <-readyLeaderCh:
			if readyMessage.Epoch == CurrentEpoch {
				if maskBit(readyMessage.ID, &leadermask) == cosi.Disabled {
					readyLeader++
					setMaskBit(readyMessage.ID, cosi.Enabled, &leadermask)
					fmt.Println(time.Now(), "ReadyLeaderCount: ", readyLeader)
				}
			}
		case <-time.After(timeoutSync * 5):
			for i := 0; i < int(gVar.ShardCnt); i++ {
				if maskBit(i, &leadermask) == cosi.Disabled && i != shard.MyMenShard.Shard {
					fmt.Println(time.Now(), "Send ReadyLeader to Shard", i, "ID", shard.ShardToGlobal[i][0])
					it = &(*ms)[shard.ShardToGlobal[i][0]]
					SendShardReadyMessage(it.Address, "reqLeaReady", readyInfo{shard.MyMenShard.Shard, CurrentEpoch})
				}
			}

		}

	}
	fmt.Println("All shards are ready.")
	//StartSendTx = make(chan bool, 1)
	//StartSendTx <- true
}

//HandleRequestShardLeaderReady handle the request from other leader
func HandleRequestShardLeaderReady(data []byte) {
	if !SentLeaderReadyFlag {
		return
	}
	data1 := make([]byte, len(data))
	copy(data1, data)
	var buff bytes.Buffer
	var payload readyInfo
	buff.Write(data1)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	if CurrentEpoch == payload.Epoch {
		it := shard.GlobalGroupMems[shard.ShardToGlobal[payload.ID][0]]
		SendShardReadyMessage(it.Address, "leaderReady", readyInfo{shard.MyMenShard.Shard, CurrentEpoch})
	}
}

//MinerReadyProcess member use this
func MinerReadyProcess() {
	var readyMessage readyInfo
	readyMessage = <-readyMemberCh
	fmt.Println("Miner waits for leader shard ready")
	for !(readyMessage.Epoch == CurrentEpoch && shard.ShardToGlobal[shard.MyMenShard.Shard][0] == readyMessage.ID) {
		readyMessage = <-readyMemberCh
	}
	SendShardReadyMessage(LeaderAddr, "shardReady", readyInfo{MyGlobalID, CurrentEpoch})
	fmt.Println(time.Now(), "Sent Ready")
}

//SendShardReadyMessage is to send shardready message
func SendShardReadyMessage(addr string, command string, message interface{}) {
	payload := gobEncode(message)
	request := append(commandToBytes(command), payload...)
	sendData(addr, request)
}

//HandleShardReady handle shard ready command
func HandleShardReady(request []byte) {
	var buff bytes.Buffer
	var payload readyInfo
	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	readyMemberCh <- payload

}

//HandleLeaderReady handle shard ready command from other leader
func HandleLeaderReady(request []byte) {
	var buff bytes.Buffer
	var payload readyInfo
	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	readyLeaderCh <- payload
}
