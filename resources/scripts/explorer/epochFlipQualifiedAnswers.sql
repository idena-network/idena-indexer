select coalesce(f.answer, '') answer, count(*) cnt
from flips f
         join transactions t on t.id = f.tx_id
         join blocks b on b.height = t.block_height
where b.epoch = $1
group by f.answer
order by f.answer