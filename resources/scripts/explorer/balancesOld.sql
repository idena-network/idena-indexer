select a.address, b.balance, b.stake
from balances b
         join addresses a on a.id = b.address_id
order by b.balance desc
limit $2 offset $1