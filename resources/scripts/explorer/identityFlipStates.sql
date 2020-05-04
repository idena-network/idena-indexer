select dfs.name status,
       count(*) cnt
from flips f
         join transactions t on t.id = f.tx_id
         join addresses a on a.id = t.from and lower(a.address) = lower($1)
    -- consider only flips with statuses
         join dic_flip_statuses dfs on dfs.id = f.status
group by dfs.name