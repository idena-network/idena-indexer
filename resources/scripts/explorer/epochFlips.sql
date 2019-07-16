select f.Cid, t.From, COALESCE(f.Status, '') status, COALESCE(short.answers, 0), COALESCE(long.answers, 0)
from flips f
         join transactions t on t.id = f.tx_id
         join blocks b on b.id = t.block_id
         join epochs e on e.id = b.epoch_id
         left join (select a.flip_id, count(*) answers from answers a where a.is_short = true group by a.flip_id) short
                   on short.flip_id = f.id
         left join (select a.flip_id, count(*) answers from answers a where a.is_short = false group by a.flip_id) long
                   on long.flip_id = f.id
where e.epoch = $1