select epoch,
       validated_count,
       block_count,
       empty_block_count,
       tx_count,
       invite_count,
       flip_count,
       burnt,
       minted,
       total_balance,
       total_stake
from epochs_detail
limit $2
offset
$1