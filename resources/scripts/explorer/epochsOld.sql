SELECT e.epoch,
       e.validation_time,
       es.validated_count,
       es.block_count,
       es.empty_block_count,
       es.tx_count,
       es.invite_count,
       es.flip_count,
       es.burnt,
       es.minted,
       es.total_balance,
       es.total_stake,
       coalesce(trew.total, 0)                  total_reward,
       coalesce(trew.validation, 0)             validation_reward,
       coalesce(trew.flips, 0)                  flips_reward,
       coalesce(trew.invitations, 0)            invitations_reward,
       coalesce(trew.foundation, 0)             foundation_payout,
       coalesce(trew.zero_wallet, 0)            zero_wallet_payout,
       coalesce(preves.min_score_for_invite, 0) min_score_for_invite
FROM epochs e
         LEFT JOIN epoch_summaries es ON es.epoch = e.epoch
         LEFT JOIN epoch_summaries preves ON preves.epoch = e.epoch - 1
         LEFT JOIN (SELECT b.epoch,
                           trew.total,
                           trew.validation,
                           trew.flips,
                           trew.invitations,
                           trew.foundation,
                           trew.zero_wallet
                    FROM total_rewards trew
                             JOIN blocks b ON b.height = trew.block_height) trew ON trew.epoch = e.epoch
ORDER BY e.epoch DESC
LIMIT $2 OFFSET $1