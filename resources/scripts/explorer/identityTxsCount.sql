select count(*) tx_count
from transactions t
         join blocks b on b.height = t.block_height
         join addresses afrom on afrom.id = t.from
         left join addresses ato on ato.id = t.to
where lower(afrom.address) = lower($1)
   or lower(ato.address) = lower($1)