select burnt,
       minted,
       total_balance,
       total_stake
from coins c
where c.block_height = $1