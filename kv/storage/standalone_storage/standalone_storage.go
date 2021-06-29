package standalone_storage

import (
	"log"
	"os"
	"path/filepath"

	"github.com/Connor1996/badger"
	"github.com/pingcap-incubator/tinykv/kv/config"
	"github.com/pingcap-incubator/tinykv/kv/storage"
	"github.com/pingcap-incubator/tinykv/kv/util/engine_util"
	"github.com/pingcap-incubator/tinykv/proto/pkg/kvrpcpb"
)

// StandAloneStorage is an implementation of `Storage` for a single-node TinyKV instance. It does not
// communicate with other nodes and all data is stored locally.
type StandAloneStorage struct {
	// Your Data Here (1).
	conf    *config.Config
	engines *engine_util.Engines
}

func NewStandAloneStorage(conf *config.Config) *StandAloneStorage {
	// Your Code Here (1).

	dbPath := conf.DBPath
	kvPath := filepath.Join(dbPath, "kv")

	os.MkdirAll(kvPath, os.ModePerm)

	kvDB := engine_util.CreateDB(kvPath, false)
	engines := engine_util.NewEngines(kvDB, nil, kvPath, "")
	return &StandAloneStorage{conf, engines}
}

func (s *StandAloneStorage) Start() error {
	// Your Code Here (1).
	return nil
}

func (s *StandAloneStorage) Stop() error {
	// Your Code Here (1).
	return s.engines.Kv.Close() //这里只能close kv, 不能调用engines自带的close
}

func (s *StandAloneStorage) Reader(ctx *kvrpcpb.Context) (storage.StorageReader, error) {
	// Your Code Here (1).
	txn := s.engines.Kv.NewTransaction(false)
	return NewStandAloneReader(txn), nil
}

func (s *StandAloneStorage) Write(ctx *kvrpcpb.Context, batch []storage.Modify) error {
	// Your Code Here (1).
	// standalone应该不用管kvrpcpb.Context
	for _, v := range batch {
		switch v.Data.(type) {
		case storage.Put:
			err := engine_util.PutCF(s.engines.Kv, v.Cf(), v.Key(), v.Value())
			if err != nil {
				return err
			}
		case storage.Delete:
			err := engine_util.DeleteCF(s.engines.Kv, v.Cf(), v.Key())
			if err != nil {
				return err
			}
		default:
			log.Fatalln("Modify's type error.")
		}
	}
	return nil
}

// StorageReader impl
type StandAloneReader struct {
	// Your code here (1)
	txn *badger.Txn
}

func NewStandAloneReader(txn *badger.Txn) *StandAloneReader {
	return &StandAloneReader{txn: txn}
}
func (s *StandAloneReader) GetCF(cf string, key []byte) ([]byte, error) {
	// Your Code here (1)
	item, err := engine_util.GetCFFromTxn(s.txn, cf, key)
	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	return item, err
}
func (s *StandAloneReader) IterCF(cf string) engine_util.DBIterator {
	// Your code here (1)
	return engine_util.NewCFIterator(cf, s.txn)
}
func (s *StandAloneReader) Close() {
	// Your code here (1)
	s.txn.Discard()
}
