select e.validation_time,
       coalesce((select b.height
                 from blocks b
                          join block_flags bf on bf.block_id = b.id
                 where b.epoch_id = e.id
                   and bf.flag = 'FlipLotteryStarted'
                ), 0)                                firstBlockHeight,
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
where e.epoch = $1