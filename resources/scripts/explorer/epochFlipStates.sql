select coalesce(dfs.name, '') status, count(*) cnt
from flips f
         join transactions t on t.id = f.tx_id
         join blocks b on b.height = t.block_height
         left join dic_flip_statuses dfs on dfs.id = f.status
where b.epoch = $1
group by dfs.name
order by dfs.name