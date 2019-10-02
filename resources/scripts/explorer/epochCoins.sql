select burnt,
       minted,
       total_balance,
       total_stake
from epochs_detail es
where es.epoch = $1