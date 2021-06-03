package gateway


// A Checkpoint stores checkpointer information to persist current block and transactions.
// Instances are created using factory methods on the implementing objects.
type Checkpoint struct {
	store CheckpointStore
}

//  Get current Block number stored in a checkpoint
func (cp *Checkpoint) GetBlockNumber() uint64 {
	return cp.store.GetBlockNumber()
}

// Set Block number to store in a checkpoint
//  Parameters:
//  blockNumber specifies the number of the block to be stored.
func (cp *Checkpoint) SetBlockNumber(blockNumber uint64)  {
	cp.store.SetBlockNumber(blockNumber)
}


//  Get current Transactions ids stored in a checkpoint
func (cp *Checkpoint) GetTransactionIds() []string {
	return cp.store.GetTransactionIds()
}


// Add transaction Id to store in an array of transactions of a checkpoint
//  Parameters:
//  transactionId specifies id of the transaction to be stored.
func (cp *Checkpoint) AddTransactionId(transactionId string) {
	cp.store.AddTransactionId(transactionId)
}

