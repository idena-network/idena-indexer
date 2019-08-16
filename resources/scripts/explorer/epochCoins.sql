select burnt_balance,
       minted_balance,
       total_balance,
       burnt_stake,
       minted_stake,
       total_stake
from epochs_detail es
where es.epoch = $1