select coalesce(sum(burnt), 0)  burnt,
       coalesce(sum(minted), 0) minted
from epochs_detail