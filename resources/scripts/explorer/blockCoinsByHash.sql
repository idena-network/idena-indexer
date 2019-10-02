select burnt,
       minted,
       total_balance,
       total_stake
from coins c
         join blocks b on b.height = c.block_height
where lower(b.hash) = lower($1)