package rccache

import (
	"fmt"

	newrep "github.com/uchihatmtkinu/RC2/NewRep"
	"github.com/uchihatmtkinu/RC2/gVar"
)

//MakeRepMsg generate the newest RepMsg
func (d *DbRef) MakeRepMsg(round uint32) newrep.RepMsg {
	mydata := new(newrep.RepMsg)
	tmpVote := make([]newrep.NewRep, gVar.ShardSize)
	for i := 0; i < len(tmpVote); i++ {
		tmpVote[i].ID = uint32(i)
		xx := uint32(0)
		if int(round) > gVar.SlidingWindows {
			xx = round - uint32(gVar.SlidingWindows)
		}
		for j := xx; j < round; j++ {
			tmpVote[i].Rep += d.RepVote[j][i].Rep
		}
	}
	mydata.Make(d.ID, *d.TBCache, tmpVote, round, &d.prk)
	d.RepFirMsg[round][d.ID] = *mydata
	d.RepFirSig[round][d.ID] = true
	return *mydata
}

//MakeRepSecMsg generate the newest RepMsg
func (d *DbRef) MakeRepSecMsg(round uint32, g newrep.GossipFirMsg) newrep.RepSecMsg {
	mydata := new(newrep.RepSecMsg)
	mydata.Make(d.ID, g, round, &d.prk)
	d.RepSecMsg[round][d.ID] = *mydata
	d.RepSecSig[round][d.ID] = true
	return *mydata
}

//GenerateGossipFir gives the data for gossip
func (d *DbRef) GenerateGossipFir(round uint32) (*newrep.GossipFirMsg, int) {
	if d.RepFirSig[round][d.ID] == false {
		d.MakeRepMsg(round)
	}
	tmp := new(newrep.GossipFirMsg)
	tmp.ID = d.ID
	tmp.Cnt = 0
	tmp.Data = nil
	cnt := 0
	tmpArr := make([]uint32, 0, gVar.ShardSize)
	for i := uint32(0); i < gVar.ShardSize; i++ {
		if d.RepFirSig[round][i] {
			tmp.Add(d.RepFirMsg[round][i])
		} else {
			cnt++
			tmpArr = append(tmpArr, i)
		}
	}
	if cnt > 0 {
		ran := d.GetGossipID(1, tmpArr)
		return tmp, int(ran)
	}
	return nil, 0

}

//GenerateGossipSec gives the data for gossip
func (d *DbRef) GenerateGossipSec(round uint32) (*newrep.GossipSecMsg, int) {
	if d.RepSecSig[round][d.ID] == false {
		tmpGossip := new(newrep.GossipFirMsg)
		tmpGossip.ID = d.ID
		tmpGossip.Cnt = 0
		tmpGossip.Data = nil
		for i := uint32(0); i < gVar.ShardSize; i++ {
			if d.RepFirSig[round][i] {
				tmpGossip.Add(d.RepFirMsg[round][i])
			}
		}
		d.MakeRepSecMsg(round, *tmpGossip)
	}

	tmp := new(newrep.GossipSecMsg)
	tmp.ID = d.ID
	tmp.Cnt = 0
	tmp.Data = nil
	cnt := 0
	tmpArr := make([]uint32, 0, gVar.ShardSize)
	for i := uint32(0); i < gVar.ShardSize; i++ {
		if d.RepSecSig[round][i] {
			tmp.Add(d.RepSecMsg[round][i])
		} else {
			cnt++
			tmpArr = append(tmpArr, i)
		}
	}
	if cnt > 0 {
		ran := d.GetGossipID(1, tmpArr)
		return tmp, int(ran)
	}
	return nil, 0
}

//UpdateGossipFir with the incoming data
func (d *DbRef) UpdateGossipFir(data newrep.GossipFirMsg) newrep.GossipFirMsg {
	d.RepVote[d.RepRound][data.ID].Rep++
	tmpRound := data.Data[0].Round
	tmp := new(newrep.GossipFirMsg)
	tmp.ID = d.ID
	tmp.Cnt = 0
	tmp.Data = nil
	if tmpRound > gVar.NumNewRep {
		fmt.Println("Data round over limit")
		return *tmp
	}
	if d.RepFirSig[tmpRound][d.ID] == false {
		d.MakeRepMsg(tmpRound)
	}
	tmpArr := make([]bool, gVar.ShardSize)
	for i := uint32(0); i < gVar.ShardSize; i++ {
		tmpArr[i] = d.RepFirSig[tmpRound][i]
	}
	for i := uint32(0); i < data.Cnt; i++ {
		tmpInShardID := data.Data[i].ID
		tmpArr[tmpInShardID] = false
		if d.RepFirSig[tmpRound][tmpInShardID] {
			if d.RepFirMsg[tmpRound][tmpInShardID].Hash() != data.Data[i].Hash() {
				d.RepByz[tmpRound][tmpInShardID] = true
			}
		} else {
			d.RepFirSig[tmpRound][tmpInShardID] = true
			d.RepFirMsg[tmpRound][tmpInShardID] = data.Data[i]
		}
	}

	for i := uint32(0); i < gVar.ShardSize; i++ {
		if tmpArr[i] {
			tmp.Add(d.RepFirMsg[tmpRound][i])
		}
	}
	//tmp.Print()
	return *tmp
}

//UpdateGossipSec with the incoming data
func (d *DbRef) UpdateGossipSec(data newrep.GossipSecMsg) (newrep.GossipSecMsg, int) {
	d.RepVote[d.RepRound][data.ID].Rep++
	tmpRound := data.Data[0].Round
	tmp := new(newrep.GossipSecMsg)
	tmp.ID = d.ID
	tmp.Cnt = 0
	tmp.Data = nil
	if tmpRound > gVar.NumNewRep {
		fmt.Println("Data round over limit")
		return *tmp, -1
	}

	tmpArr := make([]bool, gVar.ShardSize)
	for i := uint32(0); i < gVar.ShardSize; i++ {
		tmpArr[i] = d.RepSecSig[tmpRound][i]
	}
	for i := uint32(0); i < data.Cnt; i++ {
		tmpInShardID := data.Data[i].ID
		tmpArr[tmpInShardID] = false
		if !d.RepSecSig[tmpRound][tmpInShardID] {
			d.RepSecSig[tmpRound][tmpInShardID] = true
			d.RepSecMsg[tmpRound][tmpInShardID] = data.Data[i]
		}
	}
	for i := uint32(0); i < gVar.ShardSize; i++ {
		if tmpArr[i] {
			tmp.Add(d.RepSecMsg[tmpRound][i])
		}
	}
	state := 0
	if tmpArr[d.ID] {
		state = 1
	}
	return *tmp, state
}

//GetRepBlock gives the reputation block
func (d *DbRef) GetRepBlock(round uint32) *newrep.RepBlock {
	var tmp [32]byte
	if round > 0 {
		tmp = d.RepHash[round-1]
	}
	xxx := newrep.RepBlockMake(d.RepFirMsg[round][:], d.RepSecMsg[round][:], tmp)
	d.RepHash[round] = xxx.Hash
	return xxx
}
