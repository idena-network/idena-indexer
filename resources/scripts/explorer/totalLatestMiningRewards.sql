select coalesce(a.address, identities.address) address,
       coalesce(tlr.balance, 0)                balance,
       coalesce(tlr.stake, 0)                  stake,
       coalesce(proposer, 0)                   proposer,
       coalesce(final_committee, 0)            final_committee
from (select lr.address_id,
             sum(lr.balance)      balance,
             sum(lr.stake)        stake,
             sum(proposer)        proposer,
             sum(final_committee) final_committee
      from (select mr.address_id,
                   coalesce(mr.balance, 0)                   balance,
                   coalesce(mr.stake, 0)                     stake,
                   (case when mr.proposer then 1 else 0 end) proposer,
                   (case when mr.proposer then 0 else 1 end) final_committee
            from mining_rewards mr
                     join blocks b on b.height = mr.block_height
            where b."timestamp" > $1) lr
      group by lr.address_id) tlr
         join addresses a on a.id = tlr.address_id
         full outer join (select a.address, a.id
                          from address_states s
                                   join addresses a on a.id = s.address_id
                          where is_actual
                            -- 'Verified', 'Newbie', 'Human'
                            and "state" in (3, 7, 8)) identities on identities.id = tlr.address_id
order by tlr.balance + tlr.stake desc nulls last
limit $3
offset
$2