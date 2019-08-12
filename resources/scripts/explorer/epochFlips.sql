select f.Cid,
       f.Size,
       a.address              author,
       COALESCE(f.Status, '') status,
       COALESCE(f.Answer, '') answer,
       COALESCE(short.answers, 0),
       COALESCE(long.answers, 0),
       b.timestamp
from flips f
         join transactions t on t.id = f.tx_id
         join addresses a on a.id = t.from
         join blocks b on b.height = t.block_height
         left join (select a.flip_id, count(*) answers from answers a where a.is_short = true group by a.flip_id) short
                   on short.flip_id = f.id
         left join (select a.flip_id, count(*) answers from answers a where a.is_short = false group by a.flip_id) long
                   on long.flip_id = f.id
where b.epoch = $1
order by b.height desc
limit $3
    offset $2