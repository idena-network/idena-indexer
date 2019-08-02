select coalesce(f.answer, '') answer, count(*) cnt
from flips f
         join transactions t on t.id = f.tx_id
         join blocks b on b.id = t.block_id
         join epochs e on e.id = b.epoch_id
where e.epoch = $1
group by f.answer
order by f.answer