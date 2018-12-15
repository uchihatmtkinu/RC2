package network

import (
	"fmt"
	"time"

	"github.com/uchihatmtkinu/RC/NewRep"

	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

//RepGossipLoop is the loop of New reputation protocol
func RepGossipLoop(ms *[]shard.MemShard, minute int) {
	now := time.Now()
	next := now.Add(0)
	next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), minute, 0, 0, next.Location())
	tt := time.NewTimer(next.Sub(now))
	<-tt.C
	for i := uint32(0); i < gVar.NumNewRep; i++ {
		go NewRepProcess(ms, i)
		time.Sleep(time.Second * 60)
	}
	fmt.Println("Rep Donw")
	//startSync <- true
}

//NewRepProcess with gossip
func NewRepProcess(ms *[]shard.MemShard, round uint32) {
	fmt.Println("RepProcess Round", round)
	for i := uint32(0); i < gVar.ShardSize; i++ {
		CacheDbRef.RepByz[round][i] = false
		CacheDbRef.RepFirSig[round][i] = false
		CacheDbRef.RepSecSig[round][i] = false
	}
	timeoutflag := true
	for timeoutflag {
		select {
		case <-time.After(timeoutTxGap):
			CacheDbRef.Mu.Lock()
			gFM, dest := CacheDbRef.GenerateGossipFir(round)
			CacheDbRef.Mu.Unlock()
			if gFM == nil {
				break
			}
			data := new([]byte)
			gFM.Encode(data)
			go sendTxMessage(shard.GlobalGroupMems[dest].Address, "GossipFirSend", *data)
		case <-time.After(timeoutNewRep):
			fmt.Println("First NewRep Gossip reaches delay")
			timeoutflag = false
		}
	}

	timeoutflag = true
	for timeoutflag {
		select {
		case <-time.After(timeoutTxGap):
			CacheDbRef.Mu.Lock()
			sFM, dest := CacheDbRef.GenerateGossipSec(round)
			CacheDbRef.Mu.Unlock()
			if sFM == nil {
				break
			}
			data := new([]byte)
			sFM.Encode(data)
			go sendTxMessage(shard.GlobalGroupMems[dest].Address, "GossipSecSend", *data)
		case <-time.After(timeoutNewRep):
			fmt.Println("Second NewRep Gossip reaches delay")
			timeoutflag = false
		}
	}
	fmt.Println("RepProcess Round", round, " End")
	xxx := CacheDbRef.GetRepBlock(round)
	fmt.Println("RepBlock", xxx.Round, " Hash ", xxx.Hash)
	fmt.Println(xxx.Rep)
	//Generate reputation block
}

//HandleGossipFirSend handles the firSend data
func HandleGossipFirSend(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(newrep.GossipFirMsg)
	err := tmp.Decode(&data1)
	if err != nil {
		fmt.Println("Error in decoding GossipFirSend", err)
		return err
	}
	CacheDbRef.Mu.Lock()
	tmpRev := CacheDbRef.UpdateGossipFir(*tmp)
	CacheDbRef.Mu.Unlock()
	dataSend := new([]byte)
	tmpRev.Encode(dataSend)
	sendTxMessage(shard.GlobalGroupMems[tmp.ID].Address, "GossipFirRev", *dataSend)
	return nil
}

//HandleGossipFirRev handles the firSend data
func HandleGossipFirRev(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(newrep.GossipFirMsg)
	err := tmp.Decode(&data1)
	if err != nil {
		fmt.Println("Error in decoding GossipFirSend", err)
		return err
	}
	CacheDbRef.Mu.Lock()
	CacheDbRef.UpdateGossipFir(*tmp)
	CacheDbRef.Mu.Unlock()
	return nil
}

//HandleGossipSecSend handles the firSend data
func HandleGossipSecSend(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(newrep.GossipSecMsg)
	err := tmp.Decode(&data1)
	if err != nil {
		fmt.Println("Error in decoding GossipFirSend", err)
		return err
	}
	CacheDbRef.Mu.Lock()
	tmpRev := CacheDbRef.UpdateGossipSec(*tmp)
	CacheDbRef.Mu.Unlock()
	dataSend := new([]byte)
	tmpRev.Encode(dataSend)
	sendTxMessage(shard.GlobalGroupMems[tmp.ID].Address, "GossipSecRev", *dataSend)
	return nil
}

//HandleGossipSecRev handles the firSend data
func HandleGossipSecRev(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(newrep.GossipSecMsg)
	err := tmp.Decode(&data1)
	if err != nil {
		fmt.Println("Error in decoding GossipFirSend", err)
		return err
	}
	CacheDbRef.Mu.Lock()
	CacheDbRef.UpdateGossipSec(*tmp)
	CacheDbRef.Mu.Unlock()
	return nil
}
