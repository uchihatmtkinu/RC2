package network

import (
	"fmt"
	"time"

	"github.com/uchihatmtkinu/RC2/NewRep"

	"github.com/uchihatmtkinu/RC2/gVar"
	"github.com/uchihatmtkinu/RC2/shard"
)

//RepGossipLoop is the loop of New reputation protocol
func RepGossipLoop(ms *[]shard.MemShard, minute int) {
	now := time.Now()
	next := now.Add(0)
	next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), minute+1, 0, 0, next.Location())
	tt := time.NewTimer(next.Sub(now))
	<-tt.C
	fmt.Println(time.Now(), "Start Gossip Rep")
	for i := uint32(0); i < gVar.NumNewRep; i++ {
		go NewRepProcess(ms, i)
		time.Sleep(time.Second * 60)
	}
	fmt.Println(time.Now(), "Rep Done")
	//startSync <- true
}

//NewRepProcess with gossip
func NewRepProcess(ms *[]shard.MemShard, round uint32) {
	fmt.Println(time.Now(), "RepProcess Round", round)
	for i := uint32(0); i < gVar.ShardSize; i++ {
		CacheDbRef.RepByz[round][i] = false
		CacheDbRef.RepFirSig[round][i] = false
		CacheDbRef.RepSecSig[round][i] = false
	}
	timeoutflag := true
	for timeoutflag {
		select {
		case <-time.After(time.Second):
			CacheDbRef.Mu.Lock()
			gFM, dest := CacheDbRef.GenerateGossipFir(round)
			CacheDbRef.Mu.Unlock()
			if gFM == nil {
				timeoutflag = false
			} else {
				fmt.Print(time.Now(), " Round ", gFM.Data[0].Round, " GossipFirMsg to ", dest, "(", gFM.Cnt, ":")
				for i := uint32(0); i < gFM.Cnt; i++ {
					fmt.Print(gFM.Data[i].ID, ",")
				}
				fmt.Println(")")
				data := new([]byte)
				gFM.Encode(data)
				go sendTxMessage(shard.GlobalGroupMems[dest].Address, "GossipFirSend", *data)
			}
		case <-time.After(timeoutNewRep):
			fmt.Println(time.Now(), "First NewRep Gossip Time Out")
			timeoutflag = false
		}
	}
	fmt.Println(time.Now(), "Round", round, "First Rep Gossip Finished")
	timeoutflag = true
	for timeoutflag {
		select {
		case <-time.After(time.Second):
			CacheDbRef.Mu.Lock()
			sFM, dest := CacheDbRef.GenerateGossipSec(round)
			CacheDbRef.Mu.Unlock()
			if sFM == nil {
				timeoutflag = false
			} else {
				data := new([]byte)
				sFM.Encode(data)
				fmt.Print(time.Now(), " Round ", sFM.Data[0].Round, " GossipSecMsg to ", dest, "(", sFM.Cnt, ":")
				for i := uint32(0); i < sFM.Cnt; i++ {
					fmt.Print(sFM.Data[i].ID, ",")
				}
				fmt.Println(")")
				go sendTxMessage(shard.GlobalGroupMems[dest].Address, "GossipSecSend", *data)
			}
		case <-time.After(timeoutNewRep):
			fmt.Println(time.Now(), "Second NewRep Gossip Timeout")
			timeoutflag = false
		}
	}
	fmt.Println(time.Now(), "RepProcess Round", round, " End")
	CacheDbRef.Mu.Lock()
	xxx := CacheDbRef.GetRepBlock(round)
	CacheDbRef.Mu.Unlock()
	fmt.Println("RepBlock", xxx.Round, " Hash ", xxx.Hash)
	fmt.Println(xxx.Rep)
	//Generate reputation block
}

//HandleGossipFirSend handles the firSend data
func HandleGossipFirSend(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(newrep.GossipFirMsg)
	//fmt.Println("GossipFirMsg Data length:", len(data1))
	err := tmp.Decode(&data1)
	if err != nil {
		fmt.Println("Error in decoding GossipFirSend", err)
		return err
	}
	fmt.Println(time.Now(), "Round", tmp.Data[0].Round, "Get GossipFirMsg from", tmp.ID)
	CacheDbRef.Mu.Lock()
	tmpRev := CacheDbRef.UpdateGossipFir(*tmp)
	CacheDbRef.Mu.Unlock()
	fmt.Println(time.Now(), "Round", tmpRev.Data[0].Round, "GossipFirMsg Reply to", tmp.ID)
	tmpRev.Print()
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
	fmt.Println(time.Now(), "Get GossipFirMsg Reply from", tmp.ID, "Round", tmp.Data[0].Round)
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
	fmt.Print(time.Now(), " Round ", tmp.Data[0].Round, " Get GossipSecMsg from ", tmp.ID, "(", tmp.Cnt, ":")
	for i := uint32(0); i < tmp.Cnt; i++ {
		fmt.Print(tmp.Data[i].ID, ",")
	}
	fmt.Println(")")
	tmpRev, state := CacheDbRef.UpdateGossipSec(*tmp)
	CacheDbRef.Mu.Unlock()
	if state == 0 {
		fmt.Println("Not ready")
		for !CacheDbRef.RepSecSig[tmp.Data[0].Round][CacheDbRef.ID] {
			time.Sleep(200 * time.Millisecond)
		}
		CacheDbRef.Mu.Lock()
		tmpRev, _ = CacheDbRef.UpdateGossipSec(*tmp)
		CacheDbRef.Mu.Unlock()
	}
	if tmpRev.Cnt > 0 {
		fmt.Print(time.Now(), " Round ", tmp.Data[0].Round, " GossipSecMsg Reply to ", tmp.ID, "(", tmpRev.Cnt, ":")
		for i := uint32(0); i < tmpRev.Cnt; i++ {
			fmt.Print(tmpRev.Data[i].ID, ",")
		}
		fmt.Println(")")
		dataSend := new([]byte)
		tmpRev.Encode(dataSend)
		sendTxMessage(shard.GlobalGroupMems[tmp.ID].Address, "GossipSecRev", *dataSend)
	}
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
	fmt.Print(time.Now(), " Round", tmp.Data[0].Round, " Get GossipSecMsg Reply from ", tmp.ID, "(", tmp.Cnt, ":")
	for i := uint32(0); i < tmp.Cnt; i++ {
		fmt.Print(tmp.Data[i].ID, ",")
	}
	fmt.Println(")")
	CacheDbRef.Mu.Lock()
	CacheDbRef.UpdateGossipSec(*tmp)
	CacheDbRef.Mu.Unlock()
	return nil
}
