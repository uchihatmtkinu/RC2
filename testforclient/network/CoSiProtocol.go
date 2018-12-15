package network

import (
	"fmt"
	"time"

	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/Reputation/cosi"
	"github.com/uchihatmtkinu/RC/base58"
	"github.com/uchihatmtkinu/RC/ed25519"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

func LeaderCosiProcess(ms *[]shard.MemShard) cosi.SignaturePart {
	var myCommit cosi.Commitment
	var mySecret *cosi.Secret
	//var sbMessage []byte
	var it *shard.MemShard
	var cosimask []byte
	var responsemask []byte
	var commits []cosi.Commitment
	var pubKeys []ed25519.PublicKey
	var sigParts []cosi.SignaturePart

	// cosi begin

	fmt.Println(time.Now(), "Leader CoSi")

	Reputation.CurrentRepBlock.Mu.Lock()
	Reputation.CurrentRepBlock.Round++
	currentRepRound := Reputation.CurrentRepBlock.Round
	Reputation.CurrentRepBlock.Mu.Unlock()

	CoSiFlag = true
	//To simplify the problem, we just validate the previous repblock hash
	Reputation.CurrentRepBlock.Mu.RLock()
	announceMessage := announceInfo{MyGlobalID, Reputation.CurrentRepBlock.Block.Hash[:], currentRepRound, CurrentEpoch}
	Reputation.CurrentRepBlock.Mu.RUnlock()

	commits = make([]cosi.Commitment, int(gVar.ShardSize))
	pubKeys = make([]ed25519.PublicKey, int(gVar.ShardSize))
	//priKeys := make([]ed25519.PrivateKey, int(gVar.ShardSize))

	myCommit, mySecret, _ = cosi.Commit(nil)

	//byte mask 0-7 bit in one byte represent user 0-7, 8-15...
	//cosimask used in cosi announce, indicate the number of users sign the block.
	//responsemask, used in cosi, leader resent the order to the member have signed the block
	intilizeMaskBit(&cosimask, (int(gVar.ShardSize)+7)>>3, cosi.Disabled)
	intilizeMaskBit(&responsemask, (int(gVar.ShardSize)+7)>>3, cosi.Disabled)

	//handle leader's commit
	cosiCommitCh = make(chan commitInfo, bufferSize)
	commits[shard.MyMenShard.InShardId] = myCommit
	setMaskBit(shard.MyMenShard.InShardId, cosi.Enabled, &cosimask)
	setMaskBit(shard.MyMenShard.InShardId, cosi.Enabled, &responsemask)

	//sent announcement
	for i := 0; i < int(gVar.ShardSize); i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		pubKeys[it.InShardId] = it.CosiPub
		//priKeys[it.InShardId] = it.RealAccount.CosiPri
		if i != shard.MyMenShard.InShardId {
			SendCosiMessage(it.Address, "cosiAnnoun", announceMessage)
		}
	}
	fmt.Println(time.Now(), "sent CoSi announce")

	//handle members' commits
	signCount := 1
	timeoutflag := true
	cnt := 0
	for timeoutflag && signCount < int(gVar.ShardSize) {
		select {
		case commitMessage := <-cosiCommitCh:

			if commitMessage.Round == currentRepRound {
				commits[(*ms)[commitMessage.ID].InShardId] = commitMessage.Commit
				setMaskBit((*ms)[commitMessage.ID].InShardId, cosi.Enabled, &cosimask)
				signCount++
				//fmt.Println(time.Now(), "Received commit from Global ID: ", commitMessage.ID, ", commits count:", signCount, "/", int(gVar.ShardSize))
			}
		case <-time.After(timeoutCosi):
			//resend after 20 seconds
			for i := uint32(0); i < gVar.ShardSize; i++ {
				it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
				if maskBit(it.InShardId, &cosimask) == cosi.Disabled {
					fmt.Println(time.Now(), "Resend Cosi Message to", shard.ShardToGlobal[shard.MyMenShard.Shard][i])
					SendCosiMessage(it.Address, "cosiAnnoun", announceMessage)
				}
			}
			cnt++
			if cnt == 5 {
				timeoutflag = false
			}
		}
	}
	fmt.Println(time.Now(), "Recived CoSi comit")

	//fmt.Println((*ms)[GlobalAddrMapToInd[shard.MyMenShard.Address]].InShardId)

	// The leader then combines these into an aggregate commitment.
	cosigners := cosi.NewCosigners(pubKeys, cosimask)
	aggregatePublicKey := cosigners.AggregatePublicKey()
	aggregateCommit := cosigners.AggregateCommit(commits[:])

	currentChaMessage := challengeInfo{aggregatePublicKey, aggregateCommit, currentRepRound, CurrentEpoch}

	//sign or challenge
	cosiResponseCh = make(chan responseInfo, bufferSize)
	for i := uint32(0); i < gVar.ShardSize; i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		if maskBit(it.InShardId, &cosimask) == cosi.Enabled && i != uint32(shard.MyMenShard.InShardId) {
			SendCosiMessage(it.Address, "cosiChallen", currentChaMessage)
		}
	}
	fmt.Println(time.Now(), "Sent CoSi Challenage")
	//handle response
	sigParts = make([]cosi.SignaturePart, int(gVar.ShardSize))

	responseCount := 1
	//timeoutflag = true
	for responseCount < signCount {
		select {
		case reponseMessage := <-cosiResponseCh:
			if reponseMessage.Round == currentRepRound {
				it = &(*ms)[reponseMessage.ID]
				sigParts[it.InShardId] = reponseMessage.Sig
				setMaskBit(it.InShardId, cosi.Enabled, &responsemask)
				responseCount++
				//fmt.Println(time.Now(), "Received response from Global ID: ", reponseMessage.ID, ", reponses count:", responseCount, "/", signCount)
			}
		case <-time.After(timeoutCosi):
			//resend after 20 seconds
			for i := uint32(0); i < gVar.ShardSize; i++ {
				it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
				if maskBit(it.InShardId, &responsemask) == cosi.Disabled {
					fmt.Println(time.Now(), "Resend Cosi Sig to", shard.ShardToGlobal[shard.MyMenShard.Shard][i])
					SendCosiMessage(it.Address, "cosiChallen", currentChaMessage)
				}
			}
			//case <- time.After(timeoutResponse):
			//	timeoutflag = false
		}
	}
	mySigPart := cosi.Cosign(shard.MyMenShard.RealAccount.CosiPri, mySecret, announceMessage.Message, aggregatePublicKey, aggregateCommit)
	sigParts[shard.MyMenShard.InShardId] = mySigPart

	// Finally, the leader combines the two signature parts
	// into a final collective signature.
	cosiSigMessage := responseInfo{MyGlobalID, cosigners.AggregateSignature(aggregateCommit, sigParts), currentRepRound, CurrentEpoch}
	CosiData[currentRepRound+CurrentEpoch*100] = cosiSigMessage.Sig

	//currentSigMessage := cosiSigMessage{pubKeys,cosiSig}
	for i := uint32(0); i < gVar.ShardSize; i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		if maskBit(it.InShardId, &cosimask) == cosi.Enabled && i != uint32(shard.MyMenShard.InShardId) {
			SendCosiMessage(it.Address, "cosiSig", cosiSigMessage)
		}
	}

	//Add sync block
	<-FBSent
	Reputation.MyRepBlockChain.AddSyncBlock(ms, CacheDbRef.FB[CacheDbRef.ShardNum].HashID, cosiSigMessage.Sig)
	fmt.Println(time.Now(), "Add a new sync block.")
	//close CoSi
	CoSiFlag = false
	close(cosiCommitCh)
	close(cosiResponseCh)
	return cosiSigMessage.Sig
}

// MemberCosiProcess member use this
func MemberCosiProcess(ms *[]shard.MemShard) (bool, []byte) {
	//var announceMessage []byte
	// myCommit my cosi commitment
	var myCommit cosi.Commitment
	var mySecret *cosi.Secret
	var pubKeys []ed25519.PublicKey
	var it *shard.MemShard

	Reputation.CurrentRepBlock.Mu.Lock()
	Reputation.CurrentRepBlock.Round++
	currentRepRound := Reputation.CurrentRepBlock.Round
	Reputation.CurrentRepBlock.Mu.Unlock()

	//elapsed := time.Since(gVar.T1)
	//fmt.Println(time.Now(), "App elapsed: ", elapsed)
	//var timeoutflag bool
	//timeoutflag = false
	//cosiAnnounceCh = make(chan []byte)
	//elapsed := time.Since(gVar.T1)
	//
	// fmt.Println(time.Now(), "App elapsed: ", elapsed)

	cosiChallengeCh = make(chan challengeInfo)
	cosiSigCh = make(chan responseInfo)
	CoSiFlag = true
	fmt.Println(time.Now(), "Member CoSi")
	//generate pubKeys
	pubKeys = make([]ed25519.PublicKey, int(gVar.ShardSize))
	for i := 0; i < int(gVar.ShardSize); i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		pubKeys[it.InShardId] = it.CosiPub
	}

	//receive announce and verify message
	Reputation.CurrentRepBlock.Mu.RLock()
	announceMessage := Reputation.CurrentRepBlock.Block.Hash[:]
	Reputation.CurrentRepBlock.Mu.RUnlock()

	leaderAnnounceMessage := <-cosiAnnounceCh

	for leaderAnnounceMessage.Round != currentRepRound {
		leaderAnnounceMessage = <-cosiAnnounceCh
	}

	//close(cosiAnnounceCh)
	fmt.Println(time.Now(), "Leader SBM:", base58.Encode(leaderAnnounceMessage.Message))
	fmt.Println(time.Now(), "Myself SBM:", base58.Encode(announceMessage))
	if !verifySBMessage(announceMessage, leaderAnnounceMessage.Message) {
		fmt.Println("Rep Block from leader is wrong!")
		//TODO send warning
	}
	fmt.Println(time.Now(), "received cosi announce")

	//send commit
	myCommit, mySecret, _ = cosi.Commit(nil)

	commitMessage := commitInfo{MyGlobalID, myCommit, currentRepRound, CurrentEpoch}
	SendCosiMessage(LeaderAddr, "cosiCommit", commitMessage)

	fmt.Println(time.Now(), "sent cosi commit")

	//receive challenge
	//currentChaMessage := <-cosiChallengeCh
	var currentChaMessage challengeInfo
	syncFlag := true
	for syncFlag {
		select {
		case <-cosiAnnounceCh:
			fmt.Println(time.Now(), "Resend cosi commit")
			SendCosiMessage(LeaderAddr, "cosiCommit", commitMessage)
		case currentChaMessage = <-cosiChallengeCh:
			if currentChaMessage.Round == currentRepRound {
				fmt.Println(time.Now(), "received cosi challenge from leader")
				syncFlag = false
			}
		}
	}

	//send signature

	sigPart := cosi.Cosign(shard.MyMenShard.RealAccount.CosiPri, mySecret, announceMessage, currentChaMessage.AggregatePublicKey, currentChaMessage.AggregateCommit)
	responseMessage := responseInfo{MyGlobalID, sigPart, currentRepRound, CurrentEpoch}
	SendCosiMessage(LeaderAddr, "cosiRespon", responseMessage)

	//receive cosisig and verify
	var cosiSigMessage responseInfo
	syncFlag = true
	for syncFlag {
		select {
		case cosiSigMessage = <-cosiSigCh:
			if cosiSigMessage.Round == currentRepRound {
				syncFlag = false
			}
		case <-time.After(timeoutCosi):
			fmt.Println("Re-request cosi Sig")
			SendCosiMessage(LeaderAddr, "reqCosiSig", syncRequestInfo{MyGlobalID, currentRepRound, CurrentEpoch})
		}
	}

	valid := cosi.Verify(pubKeys, cosi.ThresholdPolicy(int(gVar.ShardSize)/2), announceMessage, cosiSigMessage.Sig)
	//add rep block sig
	//if valid {
	//Reputation.MyRepBlockChain.MineRepBlock(ms, CacheDbRef.FB[CacheDbRef.ShardNum].HashID, cosiSigMessage)
	<-FBSent
	Reputation.MyRepBlockChain.AddSyncBlock(ms, CacheDbRef.FB[CacheDbRef.ShardNum].HashID, cosiSigMessage.Sig)

	//}
	//close cosi
	CoSiFlag = false
	close(cosiChallengeCh)
	close(cosiSigCh)
	fmt.Println(time.Now(), "Member CoSi finished, result is ", valid)
	return valid, cosiSigMessage.Sig
}
