select *,
       (case
            when flips > 0 then (qualified + 0.5 * weakly_qualified) / flips
            else 0 end) score
from (select (select count(*)
              from flips f,
                   transactions t,
                   addresses s
              where f.tx_id = t.id
                and t.from = s.id
                and s.id = $1
                and f.status is not null) flips,

             (select count(*)
              from flips f,
                   transactions t,
                   addresses s
              where f.status = 'Qualified'
                and f.tx_id = t.id
                and t.from = s.id
                and s.id = $1)            qualified,

             (select count(*)
              from flips f,
                   transactions t,
                   addresses s
              where f.status = 'WeaklyQualified'
                and f.tx_id = t.id
                and t.from = s.id
                and s.id = $1)            weakly_qualified
     ) f