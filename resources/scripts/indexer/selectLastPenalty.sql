select p.id, p.penalty
from penalties p
         join addresses a on a.id = p.address_id
where lower(a.address) = lower($1)
  and p.id not in (select penalty_id from paid_penalties)