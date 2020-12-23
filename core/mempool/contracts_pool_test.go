package mempool

import (
	"github.com/idena-network/idena-indexer/db"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_addressContractTxs(t *testing.T) {
	holder := newAddressContractTxs()
	var txs []db.Transaction

	txs = holder.get("addr1", "contract1")
	require.Empty(t, txs)
	txs = holder.get("addr1", "contract2")
	require.Empty(t, txs)
	txs = holder.get("addr2", "contract1")
	require.Empty(t, txs)

	holder.add(&db.Transaction{
		From: "addr1",
		Hash: "1",
	}, "contract1")

	txs = holder.get("addr1", "contract1")
	require.Len(t, txs, 1)
	require.Equal(t, "1", txs[0].Hash)
	txs = holder.get("addr1", "contract2")
	require.Empty(t, txs)
	txs = holder.get("addr2", "contract1")
	require.Empty(t, txs)

	holder.add(&db.Transaction{
		From: "addr1",
		Hash: "2",
	}, "contract1")

	txs = holder.get("addr1", "contract1")
	require.Len(t, txs, 2)
	txs = holder.get("addr1", "contract2")
	require.Empty(t, txs)
	txs = holder.get("addr2", "contract1")
	require.Empty(t, txs)

	holder.add(&db.Transaction{
		From: "addr1",
		Hash: "3",
	}, "contract2")

	txs = holder.get("addr1", "contract1")
	require.Len(t, txs, 2)
	txs = holder.get("addr1", "contract2")
	require.Len(t, txs, 1)
	require.Equal(t, "3", txs[0].Hash)
	txs = holder.get("addr2", "contract1")
	require.Empty(t, txs)

	holder.add(&db.Transaction{
		From: "addr2",
		Hash: "4",
	}, "contract1")

	txs = holder.get("addr1", "contract1")
	require.Len(t, txs, 2)
	txs = holder.get("addr1", "contract2")
	require.Len(t, txs, 1)
	require.Equal(t, "3", txs[0].Hash)
	txs = holder.get("addr2", "contract1")
	require.Len(t, txs, 1)
	require.Equal(t, "4", txs[0].Hash)
	require.Len(t, holder.contractAddressesByTxHash, 4)

	holder.remove("1", "addr1")

	txs = holder.get("addr1", "contract1")
	require.Len(t, txs, 1)
	require.Equal(t, "2", txs[0].Hash)
	txs = holder.get("addr1", "contract2")
	require.Len(t, txs, 1)
	require.Equal(t, "3", txs[0].Hash)
	txs = holder.get("addr2", "contract1")
	require.Len(t, txs, 1)
	require.Equal(t, "4", txs[0].Hash)
	require.Len(t, holder.contractAddressesByTxHash, 3)

	holder.remove("2", "addr1")

	txs = holder.get("addr1", "contract1")
	require.Empty(t, txs)
	txs = holder.get("addr1", "contract2")
	require.Len(t, txs, 1)
	require.Equal(t, "3", txs[0].Hash)
	txs = holder.get("addr2", "contract1")
	require.Len(t, txs, 1)
	require.Equal(t, "4", txs[0].Hash)
	require.Len(t, holder.contractAddressesByTxHash, 2)

	holder.remove("3", "addr1")

	txs = holder.get("addr1", "contract1")
	require.Empty(t, txs)
	txs = holder.get("addr1", "contract2")
	require.Empty(t, txs)
	txs = holder.get("addr2", "contract1")
	require.Len(t, txs, 1)
	require.Equal(t, "4", txs[0].Hash)
	require.Len(t, holder.contractAddressesByTxHash, 1)

	holder.remove("4", "addr2")

	txs = holder.get("addr1", "contract1")
	require.Empty(t, txs)
	txs = holder.get("addr1", "contract2")
	require.Empty(t, txs)
	txs = holder.get("addr2", "contract1")
	require.Empty(t, txs)
	require.Empty(t, holder.contractAddressesByTxHash)
}
