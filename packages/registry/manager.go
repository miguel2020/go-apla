package registry

import (
	"sync"

	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/types"
	"github.com/AplaProject/go-apla/packages/utils"

	"github.com/AplaProject/go-apla/packages/registry/block"
	log "github.com/sirupsen/logrus"
)

type Manager struct {
	txMu      *sync.Mutex
	runningTx *MultiTransaction

	// TODO move undo log storage to file
	undoStore blockchain.LevelDBGetterPutterDeleter
}

func NewManager() *Manager {
	return &Manager{txMu: &sync.Mutex{}}
}

func (m *Manager) Begin() (*MultiTransaction, error) {
	m.txMu.Lock()
	tx, err := m.startMultiTx()
	if err != nil {
		return nil, err
	}

	m.runningTx = tx
	return tx, nil
}

func (m *Manager) Commit() error {
	rollbackOrDie := func() {
		if err := m.rollbackBlock(); err != nil {
			panic(err)
		}
	}

	if err := m.runningTx.Metadata.Commit(); err != nil {
		rollbackOrDie()
	}

	// TODO ldb, user registry dbs commiting

	m.runningTx = nil
	m.txMu.Unlock()
	return nil
}

// Rollback is reverting changes made by last block in all registries
// TODO transactional block _rollback_
func (m *Manager) RollbackBlock() error {
	m.txMu.Lock()
	defer m.txMu.Unlock()

	mtx, err := m.startMultiTx()
	if err != nil {
		return err
	}

	m.runningTx = mtx
	return m.rollbackBlock()
}

func (m *Manager) rollbackBlock() error {
	states, err := m.runningTx.log.Get()
	if err != nil {
		return err
	}

	for _, state := range states {
		// TODO blockchaindb, userregistrydb
		switch state.DBType {
		case types.DBTypeMeta:
			if err := m.runningTx.Metadata.Apply(state); err != nil {
				panic(err)
			}
		}
	}

	return nil
}

func (m *Manager) startMultiTx() (*MultiTransaction, error) {
	undo := newUndoLog(m.undoStore)

	ldbTx, err := blockchain.DB.OpenTransaction()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.LevelDBError}).Error("starting leveldb transaction")
		return nil, utils.ErrInfo(err)
	}

	dbTransaction, err := model.StartTransaction()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("starting transaction")
		return nil, utils.ErrInfo(err)
	}

	return &MultiTransaction{
		Metadata:   model.MetadataRegistry.Begin(undo),
		BlockChain: block.NewBlockRegistry(ldbTx, undo),
		log:        undo,
	}, nil
}

type MultiTransaction struct {
	Metadata      types.MetadataRegistryReaderWriter
	BlockChain    blockchain.LevelDBGetterPutterDeleter
	UsersRegistry *model.DbTransaction

	log types.StateStorage
}

func (mt *MultiTransaction) RollbackTx() error {

}
