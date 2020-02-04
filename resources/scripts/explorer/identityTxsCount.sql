select count(*) tx_count
from transactions t
where (select id from addresses where lower(address) = lower($1))
          in (t.from, t.to)