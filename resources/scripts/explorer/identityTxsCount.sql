select count(*) tx_count
from transactions t
         join blocks b on b.height = t.block_height
         join addresses afrom on afrom.id = t.from
where lower(afrom.address) = lower($1)