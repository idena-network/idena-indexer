SELECT sovcc.sort_key,
       a.address                                                  contract_address,
       autha.address                                              author,
       coalesce(b.balance, 0)                                     balance,
       ovc.fact,
       ovcs.vote_proofs,
       ovcs.votes,
       (case
            when sovcc.state = 1 then 'Open'
            when sovcc.state = 5 then 'Voted'
            when sovcc.state = 3 then 'Counting'
            when sovcc.state = 0 then 'Pending'
            when sovcc.state = 2 then 'Archive'
            when sovcc.state = 4 then 'Terminated' end)           state,
       ovcr.option,
       ovcr.votes_count                                           option_votes,
       cb.timestamp                                               create_time,
       ovc.start_time,
       head_block.height                                          head_block_height,
       head_block.timestamp                                       head_block_timestamp,
       voting_finish_b.timestamp                                  voting_finish_timestamp,
       public_voting_finish_b.timestamp                           public_voting_finish_timestamp,
       sovc.counting_block,
       coalesce(ovc.voting_min_payment, ovccs.voting_min_payment) voting_min_payment,
       ovc.quorum,
       ovc.committee_size,
       ovc.voting_duration,
       ovc.public_voting_duration,
       ovc.winner_threshold,
       ovc.owner_fee,
       true                                                       is_oracle,
       sovc.epoch,
       ovcs.finish_timestamp,
       ovcs.termination_timestamp,
       ovcs.total_reward,
       ovcs.stake
FROM (SELECT *
      FROM sorted_oracle_voting_contract_committees
      WHERE address_id = (SELECT id FROM addresses WHERE lower(address) = lower($2))
        AND ($1::text is null OR author_address_id = (SELECT id FROM addresses WHERE lower(address) = lower($1)))
        AND (
              $3::boolean AND state = 0 -- pending
              OR $4::boolean AND state = 1 -- open
              OR $5::boolean AND state = 5 -- voted
              OR $6::boolean AND state = 3 -- counting
              OR $7::boolean AND state = 2 -- completed
              OR $8::boolean AND state = 4 -- terminated
          )
        AND ($10::text IS null OR sort_key <= $10)
      ORDER BY sort_key DESC
      LIMIT $9) sovcc
         JOIN sorted_oracle_voting_contracts sovc on sovc.contract_tx_id = sovcc.contract_tx_id
         JOIN contracts c ON c.tx_id = sovcc.contract_tx_id
         JOIN addresses a on a.id = c.contract_address_id
         JOIN transactions t on t.id = sovcc.contract_tx_id
         JOIN blocks cb on cb.height = t.block_height
         JOIN addresses autha on autha.id = t.from
         LEFT JOIN balances b on b.address_id = c.contract_address_id
         JOIN oracle_voting_contracts ovc ON ovc.contract_tx_id = sovcc.contract_tx_id
         JOIN oracle_voting_contract_summaries ovcs ON ovcs.contract_tx_id = sovcc.contract_tx_id
         LEFT JOIN oracle_voting_contract_results ovcr ON ovcr.contract_tx_id = sovcc.contract_tx_id
         LEFT JOIN oracle_voting_contract_call_starts ovccs ON ovccs.ov_contract_tx_id = sovcc.contract_tx_id

         LEFT JOIN blocks voting_finish_b ON voting_finish_b.height = sovc.counting_block
         LEFT JOIN blocks public_voting_finish_b
                   ON public_voting_finish_b.height = sovc.counting_block + ovc.public_voting_duration,

     (SELECT height, timestamp FROM blocks ORDER BY height DESC LIMIT 1) head_block
ORDER BY sort_key DESC