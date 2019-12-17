select coalesce(da.name, '') answer,
       count(*) cnt
from flips f
         join transactions t on t.id = f.tx_id
         join blocks b on b.height = t.block_height
         left join dic_answers da on da.id = f.answer
where b.epoch = $1
group by da.name
order by da.name