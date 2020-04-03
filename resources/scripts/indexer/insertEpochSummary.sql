insert into epoch_summaries
(epoch,
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
 block_height,
 min_score_for_invite)
values ($1,
        (select count(*)
         from epoch_identities ei
                  join address_states s on s.id = ei.address_state_id
         where ei.epoch = $1
           -- 'Verified', 'Newbie', 'Human'
           and s.state in (3, 7, 8)),
        (select count(*)
         from blocks b
         where b.epoch = $1),
        (select count(*)
         from blocks b
         where b.epoch = $1
           and b.is_empty),
        (select count(*)
         from transactions t,
              blocks b
         where t.block_height = b.height
           and b.epoch = $1),
        (select count(*)
         from transactions t,
              blocks b
         where t.block_height = b.height
           and b.epoch = $1
           and t.type = (select id from dic_tx_types where name = 'InviteTx')),
        (select count(*)
         from flips f
                  join transactions t on t.id = f.tx_id
                  join blocks b on b.height = t.block_height and b.epoch = $1
         where f.delete_tx_id is null),
        (select coalesce(sum(burnt), 0)
         from coins c
                  join blocks b on b.height = c.block_height
         where b.epoch = $1),
        (select coalesce(sum(minted), 0)
         from coins c
                  join blocks b on b.height = c.block_height
         where b.epoch = $1),
        $3,
        $4,
        $2,
        $5)