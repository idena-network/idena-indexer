package state

import (
	"fmt"
	"github.com/cosmos/iavl"
	"github.com/golang/protobuf/proto"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/idena-network/idena-go/common/math"
	"github.com/idena-network/idena-go/core/state"
	models "github.com/idena-network/idena-go/protobuf"
	"github.com/idena-network/idena-indexer/log"
	"github.com/mholt/archiver/v3"
	"github.com/pkg/errors"
	db "github.com/tendermint/tm-db"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

type Holder interface {
	IdentityWithProof(epoch uint64, address common.Address) (*hexutil.Bytes, error)
}

func NewHolder(treeSnapshotDir string, logger log.Logger) Holder {
	return &holderImpl{
		treeSnapshotDir: treeSnapshotDir,
		statesByVersion: make(map[uint64]*state.StateDB),
		logger:          logger,
	}
}

type holderImpl struct {
	statesByVersion map[uint64]*state.StateDB
	lock            sync.RWMutex
	treeSnapshotDir string
	logger          log.Logger
}

func (h *holderImpl) IdentityWithProof(epoch uint64, address common.Address) (*hexutil.Bytes, error) {
	state, err := h.getState(epoch)
	if err != nil {
		return nil, err
	}
	valueWithProof, err := state.GetIdentityWithProof(address)
	if err != nil {
		return nil, err
	}
	if len(valueWithProof) == 0 {
		return nil, nil
	}
	res := hexutil.Bytes(valueWithProof)
	return &res, nil
}

func (h *holderImpl) getState(epoch uint64) (*state.StateDB, error) {
	h.lock.RLock()
	st, ok := h.statesByVersion[epoch]
	h.lock.RUnlock()
	if ok {
		return st, nil
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	st, ok = h.statesByVersion[epoch]
	if ok {
		return st, nil
	}
	h.logger.Info(fmt.Sprintf("Start loading state for epoch %v", epoch))
	file, err := os.Open(path.Join(h.treeSnapshotDir, fmt.Sprintf("%v.tar", epoch)))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	mdb := db.NewMemDB()
	st, err = state.NewLazy(mdb)
	if err != nil {
		return nil, err
	}
	const height uint64 = math.MaxInt64
	pdb := db.NewPrefixDB(mdb, state.StateDbKeys.BuildDbPrefix(height))
	if err := readTreeFrom(pdb, height, file); err != nil {
		return nil, err
	}
	st.CommitSnapshot(height, nil)
	h.statesByVersion[epoch] = st
	h.logger.Info(fmt.Sprintf("State for epoch %v loaded", epoch))
	return st, nil
}

func readTreeFrom(pdb *db.PrefixDB, height uint64, from io.Reader) error {
	tar := archiver.Tar{
		MkdirAll:               true,
		OverwriteExisting:      false,
		ImplicitTopLevelFolder: false,
	}

	if err := tar.Open(from, 0); err != nil {
		return err
	}

	tree := state.NewMutableTree(pdb)
	importer, err := tree.Importer(int64(height))
	if err != nil {
		return err
	}
	defer importer.Close()

	for file, err := tar.Read(); err == nil; file, err = tar.Read() {
		if data, err := ioutil.ReadAll(file); err != nil {
			common.ClearDb(pdb)
			return err
		} else {
			sb := new(models.ProtoSnapshotNodes)
			if err := proto.Unmarshal(data, sb); err != nil {
				common.ClearDb(pdb)
				return err
			}
			for _, node := range sb.Nodes {

				exportNode := &iavl.ExportNode{
					Key:     node.Key,
					Value:   node.Value,
					Version: int64(node.Version),
					Height:  int8(node.Height),
				}

				if node.EmptyValue {
					exportNode.Value = make([]byte, 0)
				}

				importer.Add(exportNode)
			}
		}
	}
	if err := importer.Commit(); err != nil {
		common.ClearDb(pdb)
		return err
	}

	if _, err := tree.LoadVersion(int64(height)); err != nil {
		common.ClearDb(pdb)
		return err
	}
	if !tree.ValidateTree() {
		common.ClearDb(pdb)
		return errors.New("corrupted tree")
	}
	return nil
}
