package newrep

import "github.com/uchihatmtkinu/RC/basic"

//NewRep is the new type of reptransaction
type NewRep struct {
	ID  uint32
	Rep uint64
}

// RepBlock reputation block
type RepBlock struct {
	PrevHash [32]byte
	Round    uint32
	TBHash   [][32]byte
	Rep      []NewRep
	Hash     [32]byte
}

//RepMsg is the first round message that transmitted by all validators
type RepMsg struct {
	ID      uint32
	Round   uint32
	NumTB   uint32
	TBHash  [][32]byte
	NumVote uint32
	Vote    []NewRep
	Sig     basic.RCSign
}

//GossipFirMsg is the first round gossip data
type GossipFirMsg struct {
	ID   uint32
	Cnt  uint32
	Data []RepMsg
}

//RepSecMsg is the second round message
type RepSecMsg struct {
	ID      uint32
	Round   uint32
	NumData uint32
	MsgHash [][32]byte
	MsgSig  []basic.RCSign
	Sig     basic.RCSign
}

//GossipSecMsg is the second round gossip data
type GossipSecMsg struct {
	ID   uint32
	Cnt  uint32
	Data []RepSecMsg
}
