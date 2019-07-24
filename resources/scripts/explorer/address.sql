select a.Address,
       coalesce(b.Balance, '0')                                 balance,
       coalesce(b.Stake, '0')                                   stake,
       (select count(*) from transactions where "from" = a.id) tx_count
from addresses a
         left join current_balances b on b.address_id = a.id
where lower(a.Address) = lower($1)