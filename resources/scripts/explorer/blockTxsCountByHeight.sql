select count(*) txs_count
from transactions t
where t.block_height = $1