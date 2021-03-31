package gateway

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	VERSION int = 1
	UNSET_BLOCK_NUMBER uint64 = -1
)

// checkpoint is the persistence state
type checkpoint struct {
	version int
	blockNumber uint64
	transactionIds []string
}

// fileCheckpoint stores checkpointer information to persist current block and transactions
// Instances are created using NewFileCheckpoint()
type fileCheckpoint struct {
	path string
	blockNumber uint64
	transactionsIds []string
}


// NewFileCheckpoint creates an instance of a checkpoint,
//  Parameters:
//  path specifies where on the filesystem to store the checkpoint.
//  Returns:
//  A Checkpoint object.
func NewFileCheckpoint(path string) (*Checkpoint, error) {
	cleanPath := filepath.Clean(path)
	transactions := make([]string,0)
	blockNumber := UNSET_BLOCK_NUMBER

	store := &fileCheckpoint{cleanPath, blockNumber, transactions }

	if _, err := os.Stat(cleanPath); os.IsExist(err) {
		// triggers if dir already exists
		store.load()
	} else {
		store.save()
	}

	return &Checkpoint{store}, nil
}

// load loads the information of block and transactions from a json format file
func (fcp *fileCheckpoint) load(){
	data := checkpoint{}
	file, _ := ioutil.ReadFile(fcp.path)
	json.Unmarshal(file, &data)
	fcp.setState(data)
}

// setState puts the information of block and transactions to the fileCheckpoint struct
func (fcp *fileCheckpoint) setState(data checkpoint) {
	fcp.blockNumber = data.blockNumber
	fcp.transactionsIds = data.transactionIds;
}

// save puts the information of block and transactions in a json file
func (fcp *fileCheckpoint) save() {
	data := checkpoint{
		version: VERSION,
		blockNumber: fcp.blockNumber,
		transactionIds: fcp.transactionsIds,
	}
	jsonData, _ := json.MarshalIndent(data, "", " ");
	_ = ioutil.WriteFile(fcp.path, jsonData, 0644)
}

// GetBlockNumber from a checkpoint
func (fcp *fileCheckpoint) GetBlockNumber() uint64 {
	return fcp.blockNumber
}

// SetBlockNumber of a checkpoint
func (fcp *fileCheckpoint) SetBlockNumber(blockNumber uint64)  {
	fcp.blockNumber = blockNumber
	fcp.transactionsIds = nil
	fcp.save()
}

// GetTransactionIds from a checkpoint
func (fcp *fileCheckpoint) GetTransactionIds() []string {
	return fcp.transactionsIds
}

// AddTransactionId to array of block transactions held in a checkpoint
func (fcp *fileCheckpoint) AddTransactionId(transactionId string) {
	fcp.transactionsIds = append(fcp.transactionsIds, transactionId)
	fcp.save()
}