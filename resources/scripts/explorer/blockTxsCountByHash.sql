select count(*) txs_count
from transactions t
         join blocks b on b.id = t.block_id
where lower(b.hash) = lower($1)