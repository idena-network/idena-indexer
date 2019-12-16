select f.cid
from flips f
         join transactions t on t.id = f.tx_id
         join blocks b on b.height = t.block_height
where b.epoch = $1