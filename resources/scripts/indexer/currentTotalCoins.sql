select total_balance, total_stake
from coins
order by block_height desc
limit 1