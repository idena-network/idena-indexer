select epoch,
       validation_time,
       validated_count,
       block_count,
       empty_block_count,
       tx_count,
       invite_count,
       flip_count,
       burnt,
       minted,
       total_balance,
       total_stake,
       total_reward,
       validation_reward,
       flips_reward,
       invitations_reward,
       foundation_payout,
       zero_wallet_payout
from epochs_detail
limit $2
offset
$1