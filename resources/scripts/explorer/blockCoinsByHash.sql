select burnt_balance,
       minted_balance,
       total_balance,
       burnt_stake,
       minted_stake,
       total_stake
from coins c
         join blocks b on b.height = c.block_height
where lower(b.hash) = lower($1)