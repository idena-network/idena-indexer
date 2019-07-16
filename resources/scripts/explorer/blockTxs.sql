select t.Hash, t.Type, b.Timestamp, t.From, t.To, t.Amount, t.Fee
from transactions t
         join blocks b on b.id = t.block_id
where b.height = $1