select count(*) flip_count
from flips f
         join transactions t on t.id = f.tx_id
         join addresses a on a.id = t.from and lower(a.address) = lower($1)
where f.delete_tx_id is null