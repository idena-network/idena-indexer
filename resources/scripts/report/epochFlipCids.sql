select f.cid
from flips f
         join transactions t on t.id = f.tx_id
         join blocks b on b.id = t.block_id
where b.epoch = $1