package rccache

import (
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

//VerifyTx verify the utxos related to transaction a
func (d *DbRef) VerifyTx(a *basic.Transaction) bool {
	for i := uint32(0); i < a.TxinCnt; i++ {
		if a.In[i].ShardIndex() == d.ShardNum {
			if !d.DB.CheckUTXO(&a.In[i], a.Hash) {
				return false
			}
		}
	}
	return true
}

//LockTx locks all utxos related to transction a
func (d *DbRef) LockTx(a *basic.Transaction) error {
	d.DB.TXCache[a.Hash] = 1
	for i := uint32(0); i < a.TxinCnt; i++ {
		if a.In[i].ShardIndex() == d.ShardNum {
			err := d.DB.LockUTXO(&a.In[i])
			if err != nil {
				log.Panic(err)
			}
		}
	}
	return nil
}

//UnlockTx locks all utxos related to transction a
func (d *DbRef) UnlockTx(a *basic.Transaction) error {
	delete(d.DB.TXCache, a.Hash)
	for i := uint32(0); i < a.TxinCnt; i++ {
		if a.In[i].ShardIndex() == d.ShardNum {
			err := d.DB.UnlockUTXO(&a.In[i])
			if err != nil {
				log.Panic(err)
			}
		}
	}
	return nil
}

//GetTx update the transaction
func (d *DbRef) GetTx(a *basic.Transaction) error {
	/*err := d.AddCache(a.Hash)
	if err != nil {
		return err
	}
	tmpPre, okw := d.WaitHashCache[basic.HashCut(a.Hash)]
	if okw {
		for i := 0; i < len(tmpPre.DataTB); i++ {
			tmpPre.StatTB[i].Valid[tmpPre.IDTB[i]] = 1
			tmpPre.StatTB[i].Stat--
			tmpPre.DataTB[i].TxArray[tmpPre.IDTB[i]] = *a
			SendingChan(&tmpPre.StatTB[i].Channel)
		}
		for i := 0; i < len(tmpPre.DataTDS); i++ {
			tmpPre.StatTDS[i].Valid[tmpPre.IDTDS[i]] = 1
			tmpPre.StatTDS[i].Stat--
			tmpPre.DataTDS[i].TxArray[tmpPre.IDTDS[i]] = a.Hash
			SendingChan(&tmpPre.StatTDS[i].Channel)
		}
		for i := 0; i < len(tmpPre.DataTL); i++ {
			tmpPre.StatTL[i].Valid[tmpPre.IDTL[i]] = 1
			tmpPre.StatTL[i].Stat--
			tmpPre.DataTL[i].TxArray[tmpPre.IDTL[i]] = a.Hash
			SendingChan(&tmpPre.StatTL[i].Channel)
		}
		//fmt.Println(time.Now(), "Sent Chan signal", base58.Encode(a.Hash[:]))
	}
	delete(d.WaitHashCache, basic.HashCut(a.Hash))
	tmp, ok := d.TXCache[a.Hash]
	if ok {
		tmp.Update(a)
	} else {
		tmp = new(CrossShardDec)
		tmp.New(a)
	}
	if tmp.InCheck[d.ShardNum] == 0 {
		if ok {
			delete(d.HashCache, basic.HashCut(a.Hash))
			delete(d.TXCache, a.Hash)
		}
		return fmt.Errorf("Not related TX")
	}
	//d.DB.AddTx(a)
	if gVar.BandDiverse {
		d.BandCnt += uint32(shard.GlobalGroupMems[d.ID].Bandwidth)
		if d.BandCnt >= gVar.MaxBand {
			tmp.Visible = true
			d.BandCnt -= gVar.MaxBand
		} else {
			tmp.Visible = false
		}
	} else {
		tmp.Visible = true
	}
	d.TXCache[a.Hash] = tmp*/
	return nil
}

//ProcessTL verify the transactions in the txlist
func (d *DbRef) ProcessTL(a *basic.TxList, tmpBatch *[]basic.TransactionBatch) error {
	d.RepVote[d.RepRound][a.Sender].Rep++
	d.TLNow = new(basic.TxDecision)
	d.TLNow.Set(d.ID, d.ShardNum, 0)
	d.TLNow.HashID = a.HashID
	d.TLNow.Single = 0
	d.TxCnt += a.TxCnt
	var tmpHash [gVar.ShardCnt][]byte
	var tmpDecision [gVar.ShardCnt]basic.TxDecision
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		tmpHash[i] = byteSlice(a.Round)
		tmpDecision[i].Set(d.ID, i, 1)
	}
	//fmt.Println(d.ID, "Process TList:")
	//a.Print()
	d.TLSCache[a.Round] = a
	*tmpBatch = make([]basic.TransactionBatch, gVar.ShardCnt)
	//tmpCnt := 0
	for i := uint32(0); i < a.TxCnt; i++ {
		/*tmp, ok := d.TXCache[a.TxArray[i]]
		if !ok {
			fmt.Println(d.ID, "Process TList: Tx ", i, "doesn't in cache", base58.Encode(a.TxArray[i][:]))
			d.TLNow.Add(0)
		} else {
			var res byte
			if tmp.InCheck[d.ShardNum] == 3 {
				if d.VerifyTx(tmp.Data) {
					res = byte(1)
					d.LockTx(tmp.Data)
				}
				if !tmp.Visible {
					res = byte(0)
				}
				if gVar.ExperimentBadLevel == 1 && d.Badness {
					res = byte(0)
				}
				d.TLNow.Add(res)
				for j := 0; j < len(tmp.ShardRelated); j++ {
					(*tmpBatch)[tmp.ShardRelated[j]].Add(tmp.Data)
					tmpHash[tmp.ShardRelated[j]] = append(tmpHash[tmp.ShardRelated[j]], a.TxArray[i][:]...)
					tmpDecision[tmp.ShardRelated[j]].Add(res)
				}
			}

		}*/
		tmpHash[0] = append(tmpHash[0], a.TxArray[i][:]...)
		tmpDecision[0].Add(1)
		d.TLNow.Add(1)
	}
	//fmt.Println("--------TxDecision from miner: ", d.ID, ":    ")
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		tmpDecision[i].HashID = sha256.Sum256(tmpHash[i])
		//fmt.Print("Decision: ", tmpDecision[i].Decision, " Sig: ")
		//fmt.Println("Hash of ", i, " : ", base58.Encode(tmpDecision[i].HashID[:]))
		tmpDecision[i].Sign(&d.prk, 0)
		//fmt.Println(tmpDecision[i].Sig[0])
		d.TLNow.Sig[i].New(&tmpDecision[i].Sig[0])
	}
	//fmt.Println(d.TLNow.Sig)
	//fmt.Println("--------TxDecision from miner: ", d.ID, " end-------------")
	d.TLRound++

	return nil
}

//GetTDS and ready to verify txblock
func (d *DbRef) GetTDS(b *basic.TxDecSet, res *[gVar.ShardSize]int64) error {
	d.RepVote[d.RepRound][b.Sender].Rep++
	d.TDSCache[b.Round] = b
	/*if d.ShardNum == b.ShardIndex {
		x, ok := d.TLSCacheMiner[b.HashID]
		if !ok {
			fmt.Print("Miner", d.ID, "doenst have such TxList")
			return fmt.Errorf("Miner %d doenst have such TxList", d.ID)
		}
		b.TxCnt = x.TxCnt
		b.TxArray = x.TxArray
	}
	index := 0
	shift := byte(0)
	for i := uint32(0); i < b.TxCnt; i++ {
		tmpHash := b.TxArray[i]
		tmp, ok := d.TXCache[tmpHash]
		tmpRes := false
		if !ok {
			tmp = new(CrossShardDec)
			tmpRes = b.Result(i)
			tmp.NewFromOther(b.ShardIndex, tmpRes)
			d.TXCache[tmpHash] = tmp
		} else {
			//fmt.Println("Result is", b.Result(i))
			tmpRes = b.Result(i)
			tmp.UpdateFromOther(b.ShardIndex, tmpRes)
			d.TXCache[tmpHash] = tmp
		}

		if b.ShardIndex == d.ShardNum {
			for j := uint32(0); j < b.MemCnt; j++ {
				tmp.Decision[shard.GlobalGroupMems[b.MemD[j].ID].InShardId] = (b.MemD[j].Decision[index]>>shift)&1 + 1
			}
			tmp.Value = 1
			if tmpRes == false {

				tmp.Decision[0] = 1
				for j := uint32(0); j < gVar.ShardSize; j++ {
					if tmp.Decision[j] == 1 {
						//shard.GlobalGroupMems[shard.ShardToGlobal[d.ShardNum][j]].Rep += gVar.RepTN * int64(tmp.Value)
						(*res)[j] += gVar.RepTN * int64(tmp.Value)
					} else if tmp.Decision[j] == 2 {
						//shard.GlobalGroupMems[shard.ShardToGlobal[d.ShardNum][j]].Rep -= gVar.RepFP * int64(tmp.Value)
						(*res)[j] -= gVar.RepFP * int64(tmp.Value)
					}
				}
			} else {
				tmp.Decision[0] = 2
				for j := uint32(0); j < gVar.ShardSize; j++ {
					if tmp.Decision[j] == 1 {
						//shard.GlobalGroupMems[shard.ShardToGlobal[d.ShardNum][j]].Rep -= gVar.RepFN * int64(tmp.Value)
						(*res)[j] -= gVar.RepFN * int64(tmp.Value)
					} else if tmp.Decision[j] == 2 {
						//shard.GlobalGroupMems[shard.ShardToGlobal[d.ShardNum][j]].Rep += gVar.RepTP * int64(tmp.Value)
						(*res)[j] += gVar.RepTP * int64(tmp.Value)
					}
				}
			}
			d.TXCache[tmpHash] = tmp
		}
		if shift < 7 {
			shift++
		} else {
			index++
			shift = 0
		}
	}
	if d.ShardNum == b.ShardIndex {
		delete(d.TLSCacheMiner, b.HashID)
	}*/
	return nil
}

//GetTxBlock handle the txblock sent by the leader
func (d *DbRef) GetTxBlock(a *basic.TxBlock) error {
	d.RepVote[d.RepRound][a.Sender].Rep++
	d.TBBCache[a.Height] = a
	/*if a.Height != d.TxB.Height+1 {
		return fmt.Errorf("Height not match")
	}
	if a.Kind != 0 {
		return fmt.Errorf("Not valid txblock type")
	}
	ok, xx := a.Verify(&shard.GlobalGroupMems[shard.ShardToGlobal[d.ShardNum][0]].RealAccount.Puk)
	if !ok {
		fmt.Println(xx)
		//return fmt.Errorf("Signature not valid")
	}
	for i := uint32(0); i < a.TxCnt; i++ {
		_, ok := d.TXCache[a.TxArray[i].Hash]
		if !ok {
			//return fmt.Errorf("Verify txblock; No tx in cache")
		}
		//if tmp.InCheckSum != 0 {
		//fmt.Println("Error in crossShard")
		//tmp.Print()
		//return fmt.Errorf("Not be fully recognized %d", i)
		//}
	}
	for i := uint32(0); i < a.TxCnt; i++ {
		d.ClearCache(a.TxArray[i].Hash)
	}
	*(d.TBCache) = append(*(d.TBCache), a.HashID)

	d.TxB = a
	d.DB.AddBlock(a)
	d.DB.UpdateUTXO(a, d.ShardNum)
	//fmt.Println("Account data of", d.ID, ": Hash", base58.Encode(a.HashID[:]))
	//d.DB.ShowAccount()*/
	return nil
}

//GetFinalTxBlock handle the txblock sent by the leader
func (d *DbRef) GetFinalTxBlock(a *basic.TxBlock) error {
	if a.Height != d.TxB.Height+1 {
		return fmt.Errorf("Height not match")
	}
	if a.Kind != 1 {
		return fmt.Errorf("Not valid txblock type")
	}
	/*ok, _ := a.Verify(&shard.GlobalGroupMems[shard.ShardToGlobal[a.ShardID][0]].RealAccount.Puk)
	if !ok {
		return fmt.Errorf("Signature not valid")
	}*/
	//*(d.TBCache) = append(*(d.TBCache), a.HashID)
	d.FB[a.ShardID] = a
	d.TxB = a
	d.DB.AddFinalBlock(a)
	return nil
}

//GetStartTxBlock handle the txblock sent by the leader
func (d *DbRef) GetStartTxBlock(a *basic.TxBlock) error {
	if a.Height != d.TxB.Height+1 {
		return fmt.Errorf("Height not match")
	}
	if a.Kind != 2 {
		return fmt.Errorf("Not valid txblock type")
	}
	ok, _ := a.Verify(&shard.GlobalGroupMems[d.Leader].RealAccount.Puk)
	if !ok {
		return fmt.Errorf("Signature not valid")
	}
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		if a.TxHash[i] != d.FB[i].HashID {
			return fmt.Errorf("final block hash not valid %d", i)
		}
	}
	d.TxB = a
	d.DB.AddStartBlock(a)
	d.DB.UploadAcc(d.ShardNum)
	return nil
}
