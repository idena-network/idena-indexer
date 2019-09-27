select count(*) flip_count
from flips f
         join transactions t on t.id = f.tx_id
         join addresses a on a.id = t.from
where lower(a.address) = lower($1)