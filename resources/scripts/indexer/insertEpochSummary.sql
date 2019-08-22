insert into epoch_summaries
(epoch,
 validated_count,
 block_count,
 empty_block_count,
 tx_count,
 invite_count,
 flip_count,
 burnt_balance,
 minted_balance,
 total_balance,
 burnt_stake,
 minted_stake,
 total_stake,
 block_height)
values ($1,
        (select count(*)
         from epoch_identities ei
                  join address_states s on s.id = ei.address_state_id
         where ei.epoch = $1
           and s.state in ('Verified', 'Newbie')),
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
           and t.type = 'InviteTx'),
        (select count(*)
         from flips f,
              transactions t,
              blocks b
         where f.tx_id = t.id
           and t.block_height = b.height
           and b.epoch = $1),
        (select coalesce(sum(burnt_balance), 0)
         from coins c
                  join blocks b on b.height = c.block_height
         where b.epoch = $1),
        (select coalesce(sum(minted_balance), 0)
         from coins c
                  join blocks b on b.height = c.block_height
         where b.epoch = $1),
        $3,
        (select coalesce(sum(burnt_stake), 0)
         from coins c
                  join blocks b on b.height = c.block_height
         where b.epoch = $1),
        (select coalesce(sum(minted_stake), 0) minted_stake
         from coins c
                  join blocks b on b.height = c.block_height
         where b.epoch = $1),
        $4,
        $2)