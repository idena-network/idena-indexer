select epoch,
       validated_count,
       block_count,
       tx_count,
       invite_count,
       flip_count,
       burnt_balance,
       minted_balance,
       total_balance,
       burnt_stake,
       minted_stake,
       total_stake
from epochs_detail
limit $2
    offset $1