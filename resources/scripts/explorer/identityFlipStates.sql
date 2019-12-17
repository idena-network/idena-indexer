select dfs.name status,
       count(*) cnt
from flips f
         join transactions t on t.id = f.tx_id
         join addresses s on s.id = t.from
    -- consider only flips with statuses
         join dic_flip_statuses dfs on dfs.id = f.status
where lower(s.address) = lower($1)
group by dfs.name