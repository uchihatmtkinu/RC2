package network

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"

	"fmt"

	"math/rand"

	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

//var currentTxList *[]basic.TxList
//var currentTxDecSet *[]basic.TxDecSet

//RepProcessLoop is the loop of reputation
func RepProcessLoop(ms *[]shard.MemShard) {
	flag := true
	for flag {
		flag = RepProcess(ms)
		time.Sleep(time.Second * 2)
	}
	fmt.Println("Start to sync")
	startSync <- true
}

//RepProcess pow proces
func RepProcess(ms *[]shard.MemShard) bool {
	var it *shard.MemShard
	var validateBlock Reputation.RepPowInfo
	var item Reputation.RepPowInfo
	Reputation.IDToNonce = make([]int, int(gVar.ShardSize))
	Reputation.NonceMap = make(map[int]int)
	Reputation.StartCalPoWAnnounce = make(chan bool)
	res := <-startRep
	fmt.Println(time.Now(), "Do reputation block")
	flag := true
	Reputation.RepPowTxCh = make(chan Reputation.RepPowInfo)
	Reputation.RepPowRxValidate = make(chan Reputation.RepPowInfo)
	tmp := res.Hash
	go Reputation.MyRepBlockChain.MineRepBlock(res.Rep, &tmp, MyGlobalID)
	for flag {
		select {
		case item = <-Reputation.RepPowTxCh:
			{
				Reputation.NonceMap[item.Nonce]++
				for i := 0; i < int(gVar.ShardSize); i++ {
					if i != shard.MyMenShard.InShardId {
						it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
						//fmt.Println("have sent " + it.Address)
						go SendRepPowMessage(it.Address, "RepPowAnnou", powInfo{MyGlobalID, item.Round, item.Hash, item.Nonce})
					}
				}
				flag = false
			}
		case validateBlock = <-Reputation.RepPowRxValidate:
			{
				Reputation.NonceMap[validateBlock.Nonce]++
				for i := 0; i < int(gVar.ShardSize); i++ {
					if i != shard.MyMenShard.InShardId {
						it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
						//fmt.Println("have sent " + it.Address)
						go SendRepPowMessage(it.Address, "RepPowAnnou", powInfo{MyGlobalID, validateBlock.Round, validateBlock.Hash, validateBlock.Nonce})
					}
				}
				flag = false
				fmt.Println("add pow rep block from others")
			}
		}
	}
	//if nonce is wrong request for reputation block
	RxRepBlockCh = make(chan *Reputation.RepBlock, 10)
	correctNonce := 0
	askrep := 0
	halfflag := true
	for key, value := range Reputation.NonceMap {
		if value >= int(gVar.ShardSize/2) {
			correctNonce = key
			halfflag = false
		}
	}
	fmt.Println(Reputation.NonceMap, halfflag)
	<-Reputation.StartCalPoWAnnounce
	for halfflag {
		item = <-Reputation.RepPowRxCh
		Reputation.CurrentRepBlock.Mu.RLock()
		if item.Round == Reputation.CurrentRepBlock.Round {
			Reputation.NonceMap[item.Nonce]++
			//fmt.Println(time.Now(), "Nonce:", item.Nonce, " value:", Reputation.NonceMap[item.Nonce])
			if Reputation.NonceMap[item.Nonce] >= int(gVar.ShardSize/2) {
				Reputation.IDToNonce[shard.GlobalGroupMems[item.ID].InShardId] = item.Nonce
				correctNonce = item.Nonce
				halfflag = false
			}
		}
		Reputation.CurrentRepBlock.Mu.RUnlock()

	}
	Reputation.CurrentRepBlock.Mu.RLock()
	if correctNonce != Reputation.CurrentRepBlock.Block.Nonce {
		fmt.Println("Rep Nounce not match")
		rand.Seed(int64(shard.MyMenShard.Shard*3000+shard.MyMenShard.InShardId) + time.Now().UTC().UnixNano())
		askrep = rand.Intn(int(gVar.ShardSize))
		for Reputation.IDToNonce[askrep] != correctNonce {
			askrep = (askrep + 1) % int(gVar.ShardSize)
		}
		SendRepPowMessage((*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][askrep]].Address, "RequestRep", requetRepInfo{MyGlobalID, Reputation.CurrentRepBlock.Round})
		receiveflag := true
		for receiveflag {
			select {
			case receiveRepBlock := <-RxRepBlockCh:
				{
					if receiveRepBlock.Nonce == correctNonce {
						Reputation.MyRepBlockChain.AddRepBlockFromOthers(receiveRepBlock)
						for _, txs := range receiveRepBlock.RepTransactions {
							(*ms)[txs.GlobalID].Rep = txs.Rep
						}
						receiveflag = false
					}
				}
			case <-time.After(timeoutSync):
				{
					askrep = (askrep + 1) % int(gVar.ShardSize)
					for Reputation.IDToNonce[askrep] != correctNonce {
						askrep = (askrep + 1) % int(gVar.ShardSize)
					}
					SendRepPowMessage((*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][askrep]].Address, "RequestRep", requetRepInfo{MyGlobalID, Reputation.CurrentRepBlock.Round})
				}
			}
		}
	}
	fmt.Println("PoW ", Reputation.CurrentRepBlock.Round, " round finished.")
	Reputation.CurrentRepBlock.Mu.RUnlock()
	close(Reputation.RepPowRxValidate)
	close(Reputation.RepPowTxCh)
	return res.Last
	//close(Reputation.RepPowRxCh)
}

// SendRepPowMessage send reputation block
func SendRepPowMessage(addr string, command string, message interface{}) {
	payload := gobEncode(message)
	request := append(commandToBytes(command), payload...)
	sendData(addr, request)
}

// HandleRepPowRx receive reputation block
func HandleRepPowRx(request []byte) {
	var buff bytes.Buffer
	var payload powInfo

	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	Reputation.CurrentRepBlock.Mu.RLock()
	if payload.Round >= Reputation.CurrentRepBlock.Round {
		Reputation.RepPowRxCh <- Reputation.RepPowInfo{payload.ID, payload.Round, payload.Nonce, payload.Hash}
		//fmt.Println(time.Now(), "Received PoW from others", payload.ID, payload.Round, payload.Nonce)
	}
	Reputation.CurrentRepBlock.Mu.RUnlock()

}

// HandleRepBlockRx receive reputation block
func HandleRepBlockRx(request []byte) {
	var buff bytes.Buffer
	//var payload RepBlockRxInfo
	var payload Reputation.RepBlock
	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	//RxRepBlockCh <- &payload.Block
	RxRepBlockCh <- &payload
	fmt.Println("Received Reputation Block from others")

}

// HandleRepBlockRx receive reputation block
func HandleRequestRepBlockRx(request []byte) {
	var buff bytes.Buffer
	var payload requetRepInfo
	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	Reputation.CurrentRepBlock.Mu.RLock()
	if payload.Round == Reputation.CurrentRepBlock.Round {
		//SendRepPowMessage(shard.GlobalGroupMems[payload.ID].Address, "RepBlock", Reputation.CurrentRepBlock.Block.Serialize())
		SendRepPowMessage(shard.GlobalGroupMems[payload.ID].Address, "RepBlock", Reputation.CurrentRepBlock.Block)
		fmt.Println("Send Reputation Block To ID: ", payload.ID)
	}
	Reputation.CurrentRepBlock.Mu.RUnlock()

}
