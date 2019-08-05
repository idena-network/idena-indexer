select f.status, count(*)
from flips f
         join transactions t on t.id = f.tx_id
         join addresses s on s.id = t.from
where lower(s.address) = lower($1)
  and f.status is not null
group by f.status