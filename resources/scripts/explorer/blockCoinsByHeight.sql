select burnt_balance,
       minted_balance,
       total_balance,
       burnt_stake,
       minted_stake,
       total_stake
from coins c
where c.block_height = $1