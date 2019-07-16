select t.Hash, t.Type, b.Timestamp, t.From, COALESCE(t.To, ''), t.Amount, t.Fee
from transactions t
         join blocks b on b.id = t.block_id
         join epochs e on e.id = b.epoch_id
where e.epoch = $1