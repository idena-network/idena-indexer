select (select count(*)
        from epoch_identities ei
                 join address_states s on s.id = ei.address_state_id
        where s.state = 'Verified'
          and ei.epoch_id = e.id)                    verified,
       (select count(*)
        from epoch_identities ei
                 join address_states s on s.id = ei.address_state_id
        where s.state != 'Verified'
          and ei.epoch_id = e.id)                    not_verified,

       (select count(*)
        from flips f
                 join transactions t on t.id = f.tx_id
                 join blocks b on b.id = t.block_id
        where b.epoch_id = e.id)                     submitted_flips,

       (select count(*)
        from flips f
                 join transactions t on t.id = f.tx_id
                 join blocks b on b.id = t.block_id
        where b.epoch_id = e.id
          and f.id in (select flip_id from answers)) solved_flips,

       (select count(*)
        from flips f
                 join transactions t on t.id = f.tx_id
                 join blocks b on b.id = t.block_id
        where b.epoch_id = e.id
          and f.status = 'Qualified')                qualified_flips,
       (select count(*)
        from flips f
                 join transactions t on t.id = f.tx_id
                 join blocks b on b.id = t.block_id
        where b.epoch_id = e.id
          and f.status = 'WeaklyQualified')          weakly_qualified_flips,
       (select count(*)
        from flips f
                 join transactions t on t.id = f.tx_id
                 join blocks b on b.id = t.block_id
        where b.epoch_id = e.id
          and f.status = 'NotQualified')             not_qualified_flips,

       (select count(*)
        from flips f
                 join transactions t on t.id = f.tx_id
                 join blocks b on b.id = t.block_id
        where b.epoch_id = e.id
          and f.answer = 'Inappropriate')            inappropriate_flips


from epochs e
where e.epoch =
      (select max(epoch) epoch
       from epochs e
                join epoch_identities ei on ei.epoch_id = e.id)