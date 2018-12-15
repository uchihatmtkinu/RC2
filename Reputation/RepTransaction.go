package Reputation

type RepTransaction struct {
	GlobalID   	int
	//AddrReal 	[32]byte //public key -> id
	Rep			int64
}

//new reputation transaction
func NewRepTransaction(globalID int, rep int64) *RepTransaction{
	tx := RepTransaction{globalID,rep/10}
	return &tx
}


// SetID sets ID of a transaction
/*
func (tx *RepTransaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}
*/