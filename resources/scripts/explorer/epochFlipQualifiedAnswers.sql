select coalesce(da.name, '') answer,
       count(*) cnt
from flips f
         join transactions t on t.id = f.tx_id
         join blocks b on b.height = t.block_height and b.epoch = $1
         left join dic_answers da on da.id = f.answer
where f.delete_tx_id is null
group by da.name
order by da.name