with cur_epoch as (
    select burnt, minted, total_balance, total_stake
    from (SELECT COALESCE(sum(c.minted), 0), COALESCE(sum(c.burnt), 0)
          FROM coins c
                   JOIN blocks b ON b.height = c.block_height
          WHERE b.epoch = (select max(epoch) from epochs)) t1 (minted, burnt),
         (SELECT c.total_balance, c.total_stake
          FROM coins c
                   JOIN blocks b ON b.height = c.block_height
          ORDER BY c.block_height DESC
          LIMIT 1) t2
)
SELECT e.epoch,
       e.validation_time                                                    validation_time,
       COALESCE(es.validated_count, (select count(*)
                                     from address_states
                                     where is_actual
                                       -- 'Verified', 'Newbie', 'Human'
                                       and state in (3, 7, 8)))          AS validated_count,
       COALESCE(es.block_count, (SELECT count(*) AS count
                                 FROM blocks b
                                 WHERE b.epoch = e.epoch))               AS block_count,
       COALESCE(es.empty_block_count, (SELECT count(*) AS count
                                       FROM blocks b
                                       WHERE b.epoch = e.epoch
                                         and b.is_empty))                AS empty_block_count,
       COALESCE(es.tx_count, (SELECT count(*) AS count
                              FROM transactions t,
                                   blocks b
                              WHERE t.block_height = b.height
                                AND b.epoch = e.epoch))                  AS tx_count,
       COALESCE(es.invite_count, (SELECT count(*) AS count
                                  FROM transactions t,
                                       blocks b
                                  WHERE t.block_height = b.height
                                    AND b.epoch = e.epoch
                                    AND t.type = 2))                     AS invite_count,
       COALESCE(es.flip_count, (select count(*)
                                from flips f
                                         join transactions t on t.id = f.tx_id
                                         join blocks b on b.height = t.block_height and b.epoch = e.epoch
                                where f.delete_tx_id is null))           AS flip_count,
       COALESCE(es.burnt, (select burnt from cur_epoch))                 AS burnt,
       COALESCE(es.minted, (select minted from cur_epoch))               AS minted,
       COALESCE(es.total_balance, (select total_balance from cur_epoch)) AS total_balance,
       COALESCE(es.total_stake, (select total_stake from cur_epoch))     AS total_stake,
       coalesce(trew.total, 0)                                              total_reward,
       coalesce(trew.validation, 0)                                         validation_reward,
       coalesce(trew.flips, 0)                                              flips_reward,
       coalesce(trew.invitations, 0)                                        invitations_reward,
       coalesce(trew.foundation, 0)                                         foundation_payout,
       coalesce(trew.zero_wallet, 0)                                        zero_wallet_payout,
       coalesce(preves.min_score_for_invite, 0)                             min_score_for_invite
FROM epochs e
         LEFT JOIN epoch_summaries es ON es.epoch = e.epoch
         left join epoch_summaries preves on preves.epoch = e.epoch - 1
         left join (select b.epoch,
                           trew.total,
                           trew.validation,
                           trew.flips,
                           trew.invitations,
                           trew.foundation,
                           trew.zero_wallet
                    from total_rewards trew
                             join blocks b on b.height = trew.block_height) trew on trew.epoch = e.epoch
ORDER BY e.epoch DESC
limit $2
offset
$1