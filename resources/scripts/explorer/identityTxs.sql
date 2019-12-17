select t.Hash,
       dtt.name                  "type",
       b.Timestamp,
       afrom.Address             "from",
       COALESCE(ato.Address, '') "to",
       t.Amount,
       t.Tips,
       t.max_fee,
       t.Fee,
       t.size
from transactions t
         join blocks b on b.height = t.block_height
         join addresses afrom on afrom.id = t.from
         left join addresses ato on ato.id = t.to
         join dic_tx_types dtt on dtt.id = t.Type
where lower(afrom.address) = lower($1)
   or lower(ato.address) = lower($1)
order by b.height desc
limit $3
offset
$2