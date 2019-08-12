select count(*) txs_count
from transactions t
         join blocks b on b.height = t.block_height
where lower(b.hash) = lower($1)