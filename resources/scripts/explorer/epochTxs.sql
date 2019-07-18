select t.Hash, t.Type, b.Timestamp, afrom.Address "from", COALESCE(ato.Address, '') "to", t.Amount, t.Fee
from transactions t
         join blocks b on b.id = t.block_id
         join epochs e on e.id = b.epoch_id
         join addresses afrom on afrom.id = t.from
         left join addresses ato on ato.id = t.to
where e.epoch = $1