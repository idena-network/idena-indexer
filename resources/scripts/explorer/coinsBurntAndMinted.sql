select coalesce(sum(burnt_balance), 0)  burnt_balance,
       coalesce(sum(minted_balance), 0) minted_balance,
       coalesce(sum(burnt_stake), 0)    burnt_stake,
       coalesce(sum(minted_stake), 0)   minted_stake
from epochs_detail