select count(*) txs_count
from transactions t
         join blocks b on b.id = t.block_id
where b.height = $1