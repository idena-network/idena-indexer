select burnt,
       minted,
       total_balance,
       total_stake
from epoch_summaries
where epoch = $1