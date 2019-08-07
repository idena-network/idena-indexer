insert into used_invites (invite_tx_id, activation_tx_id)
values ((
            select t.id
            from transactions t
                     join addresses a on a.id = t.to
                     join blocks b on b.id = t.block_id
            where b.epoch_id = $1
              and lower(a.address) = lower($2)
              and t.type = 'InviteTx'
            order by height desc
            limit 1
        ), $3)