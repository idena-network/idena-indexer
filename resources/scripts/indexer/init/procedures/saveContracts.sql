CREATE OR REPLACE PROCEDURE save_tx_receipts(p_items tp_tx_receipt[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item  tp_tx_receipt;
    l_tx_id bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id INTO l_tx_id FROM transactions WHERE lower(hash) = lower(l_item.tx_hash);

            INSERT INTO tx_receipts (tx_id, success, gas_used, gas_cost, "method", error_msg)
            VALUES (l_tx_id, l_item.success, l_item.gas_used, l_item.gas_cost, l_item.method, l_item.error_msg);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_oracle_voting_contracts(p_block_height bigint,
                                                         p_items tp_oracle_voting_contract[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    CONTRACT_TYPE_ORACLE_VOTING CONSTANT smallint        = 2;
    SOVC_STATE_PENDING          CONSTANT smallint        = 0;
    MAX_VOTING_MIN_PAYMENT      CONSTANT numeric(48, 18) = 999999999999999999999999999999;
    MAX_ORACLE_REWARD_FUND      CONSTANT numeric(48, 18) = MAX_VOTING_MIN_PAYMENT;
    l_item                               tp_oracle_voting_contract;
    l_contract_address_id                bigint;
    l_author_address_id                  bigint;
    l_tx_id                              bigint;
    l_estimated_oracle_reward            numeric;
    l_voting_min_payment                 numeric;
    l_refund_recipient_address_id        bigint;
    l_oracle_reward_fund                 numeric;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            l_contract_address_id = get_address_id_or_insert(p_block_height, l_item.contract_address);

            SELECT id, "from"
            INTO l_tx_id, l_author_address_id
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            INSERT INTO contracts (tx_id, contract_address_id, "type", stake)
            VALUES (l_tx_id, l_contract_address_id, CONTRACT_TYPE_ORACLE_VOTING, l_item.stake);

            l_voting_min_payment = null_if_negative_numeric(l_item.voting_min_payment);
            if l_voting_min_payment is not null and l_voting_min_payment > MAX_VOTING_MIN_PAYMENT then
                l_voting_min_payment = MAX_VOTING_MIN_PAYMENT;
            end if;

            l_oracle_reward_fund = null_if_negative_numeric(l_item.oracle_reward_fund);
            if l_oracle_reward_fund is not null and l_oracle_reward_fund > MAX_ORACLE_REWARD_FUND then
                l_oracle_reward_fund = MAX_ORACLE_REWARD_FUND;
            end if;

            l_refund_recipient_address_id = null;
            if l_item.refund_recipient is not null and l_item.refund_recipient <> '' then
                l_refund_recipient_address_id = get_address_id_or_insert(p_block_height, l_item.refund_recipient);
            end if;

            INSERT INTO oracle_voting_contracts (contract_tx_id, start_time, voting_duration,
                                                 voting_min_payment, fact, public_voting_duration, winner_threshold,
                                                 quorum, committee_size, owner_fee, state,
                                                 owner_deposit, oracle_reward_fund, refund_recipient_address_id, hash)
            VALUES (l_tx_id, l_item.start_time, l_item.voting_duration, l_voting_min_payment,
                    decode(l_item.fact, 'hex'), l_item.public_voting_duration, l_item.winner_threshold, l_item.quorum,
                    l_item.committee_size, l_item.owner_fee, l_item.state,
                    null_if_negative_numeric(l_item.owner_deposit), l_oracle_reward_fund,
                    l_refund_recipient_address_id, decode(l_item.hash, 'hex'));

            l_estimated_oracle_reward = calculate_estimated_oracle_reward(0, l_tx_id);
            INSERT INTO sorted_oracle_voting_contracts (contract_tx_id, author_address_id, sort_key, state, state_tx_id,
                                                        epoch)
            VALUES (l_tx_id, l_author_address_id,
                    calculate_oracle_voting_contract_sort_key(l_estimated_oracle_reward, l_tx_id), SOVC_STATE_PENDING,
                    l_tx_id, null);

            INSERT INTO oracle_voting_contract_summaries (contract_tx_id, vote_proofs, votes, finish_timestamp,
                                                          termination_timestamp, total_reward, stake)
            VALUES (l_tx_id, 0, 0, null, null, null, l_item.stake);

            INSERT INTO oracle_voting_contract_authors_and_open_voters (deploy_or_vote_tx_id, contract_tx_id, address_id)
            VALUES (l_tx_id, l_tx_id, l_author_address_id);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_oracle_voting_contract_call_starts(p_block_height bigint,
                                                                    p_items jsonb[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    SOVC_STATE_VOTING      CONSTANT smallint        = 1;
    MAX_VOTING_MIN_PAYMENT CONSTANT numeric(48, 18) = 999999999999999999999999999999;
    l_item                          jsonb;
    l_contract_address_id           bigint;
    l_contract_tx_id                bigint;
    l_tx_id                         bigint;
    l_start_block_height            bigint;
    l_voting_duration               bigint;
    l_epoch                         bigint;
    l_voting_min_payment            numeric;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to"
            INTO l_tx_id, l_contract_address_id
            FROM transactions
            WHERE lower(hash) = lower((l_item ->> 'txHash')::text);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            l_start_block_height = (l_item ->> 'startBlockHeight')::bigint;
            l_epoch = (l_item ->> 'epoch')::bigint;
            l_voting_min_payment = null_if_negative_numeric((l_item ->> 'votingMinPayment')::numeric);
            if l_voting_min_payment is not null and l_voting_min_payment > MAX_VOTING_MIN_PAYMENT then
                l_voting_min_payment = MAX_VOTING_MIN_PAYMENT;
            end if;
            INSERT INTO oracle_voting_contract_call_starts (call_tx_id, ov_contract_tx_id, start_block_height, epoch,
                                                            voting_min_payment, vrf_seed, state)
            VALUES (l_tx_id, l_contract_tx_id, l_start_block_height, l_epoch, l_voting_min_payment,
                    decode(l_item ->> 'vrfSeed', 'hex'), (l_item ->> 'state')::smallint);

            SELECT voting_duration
            INTO l_voting_duration
            FROM oracle_voting_contracts
            WHERE contract_tx_id = l_contract_tx_id;
            call update_sorted_oracle_voting_contracts(p_block_height, l_contract_tx_id, null, SOVC_STATE_VOTING,
                                                       l_tx_id, l_start_block_height + l_voting_duration, l_epoch);

            call save_oracle_voting_committee(p_block_height, l_contract_tx_id, l_item -> 'committee');
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_oracle_voting_contract_call_vote_proofs(p_block_height bigint,
                                                                         p_items tp_oracle_voting_contract_call_vote_proof[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    SOVCC_STATE_VOTED CONSTANT smallint = 5;
    l_item                     tp_oracle_voting_contract_call_vote_proof;
    l_tx_id                    bigint;
    l_address_id               bigint;
    l_contract_address_id      bigint;
    l_contract_tx_id           bigint;
    l_is_first                 boolean;
    l_secret_votes_count       bigint;
    l_vote_proofs_diff         bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "from", "to"
            INTO l_tx_id, l_address_id, l_contract_address_id
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            l_is_first = (SELECT not exists(SELECT 1
                                            FROM oracle_voting_contract_call_vote_proofs
                                            WHERE address_id = l_address_id
                                              AND ov_contract_tx_id = l_contract_tx_id));

            l_secret_votes_count = null_if_negative_bigint(l_item.secret_votes_count);

            INSERT INTO oracle_voting_contract_call_vote_proofs (call_tx_id, ov_contract_tx_id, address_id, vote_hash,
                                                                 secret_votes_count, discriminated)
            VALUES (l_tx_id, l_contract_tx_id, l_address_id, decode(l_item.vote_hash, 'hex'), l_secret_votes_count,
                    l_item.discriminated);

            if l_is_first then
                call update_sorted_oracle_voting_contract_committees(p_block_height, l_contract_tx_id, l_address_id,
                                                                     null, SOVCC_STATE_VOTED, null, true);
                l_vote_proofs_diff = 1;
            else
                l_vote_proofs_diff = 0;
            end if;

            if l_vote_proofs_diff > 0 or l_secret_votes_count IS NOT NULL then
                call update_oracle_voting_contract_summaries(p_block_height, l_contract_tx_id, l_vote_proofs_diff, 0,
                                                             null, null, null,
                                                             0, l_secret_votes_count, null);
            end if;
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_oracle_voting_contract_call_votes(p_block_height bigint,
                                                                   p_items tp_oracle_voting_contract_call_vote[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                 tp_oracle_voting_contract_call_vote;
    l_tx_id                bigint;
    l_contract_address_id  bigint;
    l_sender_address_id    bigint;
    l_contract_tx_id       bigint;
    l_delegatee_address_id bigint;
    l_option_all_votes     bigint;
    l_option_votes         bigint;
    l_prev_option_votes    bigint;
    l_prev_pool_vote       smallint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to", "from"
            INTO l_tx_id, l_contract_address_id, l_sender_address_id
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            if l_item.delegatee is not null then
                l_delegatee_address_id = get_address_id_or_insert(p_block_height, l_item.delegatee);
            else
                l_delegatee_address_id = null;
            end if;

            l_option_all_votes = null_if_negative_bigint(l_item.option_all_votes);
            l_option_votes = null_if_negative_bigint(l_item.option_votes);
            l_prev_option_votes = null_if_negative_bigint(l_item.prev_option_votes);
            l_prev_pool_vote = null_if_negative_bigint(l_item.prev_pool_vote)::smallint;
            INSERT INTO oracle_voting_contract_call_votes (call_tx_id, ov_contract_tx_id, vote, salt, option_votes,
                                                           option_all_votes, secret_votes_count, delegatee_address_id,
                                                           prev_pool_vote, prev_option_votes, discriminated)
            VALUES (l_tx_id, l_contract_tx_id, l_item.vote, decode(l_item.salt, 'hex'),
                    l_option_votes,
                    l_option_all_votes,
                    null_if_negative_bigint(l_item.secret_votes_count),
                    l_delegatee_address_id,
                    l_prev_pool_vote,
                    l_prev_option_votes,
                    l_item.discriminated);

            call update_oracle_voting_contract_summaries(p_block_height, l_contract_tx_id, 0, 1, null, null, null, 0,
                                                         l_item.secret_votes_count, null);

            if l_option_all_votes is null then
                -- TODO needed to support old contracts version
                call add_vote_to_oracle_voting_contract_result(p_block_height, l_contract_tx_id, l_item.vote);
            else
                call update_oracle_voting_contract_result(p_block_height, l_contract_tx_id, l_item.vote,
                                                          l_option_votes, l_option_all_votes);

                if l_prev_option_votes is not null then
                    call update_oracle_voting_contract_result(p_block_height, l_contract_tx_id, l_prev_pool_vote,
                                                              l_prev_option_votes, null);
                end if;
            end if;

            INSERT INTO oracle_voting_contract_authors_and_open_voters (deploy_or_vote_tx_id, contract_tx_id, address_id)
            VALUES (l_tx_id, l_contract_tx_id, l_sender_address_id)
            ON CONFLICT DO NOTHING;
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_oracle_voting_contract_call_finishes(p_block_height bigint,
                                                                      p_items tp_oracle_voting_contract_call_finish[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    SOVC_STATE_COMPLETED CONSTANT smallint = 2;
    l_item                        tp_oracle_voting_contract_call_finish;
    l_tx_id                       bigint;
    l_contract_address_id         bigint;
    l_contract_tx_id              bigint;
    l_block_timestamp             bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to"
            INTO l_tx_id, l_contract_address_id
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            INSERT INTO oracle_voting_contract_call_finishes (call_tx_id, ov_contract_tx_id, result, fund,
                                                              oracle_reward, owner_reward, state)
            VALUES (l_tx_id, l_contract_tx_id, null_if_negative_bigint(l_item.result)::smallint, l_item.fund,
                    l_item.oracle_reward, l_item.owner_reward, l_item.state);

            SELECt "timestamp" INTO l_block_timestamp FROM blocks WHERE height = p_block_height;
            call update_oracle_voting_contract_summaries(p_block_height, l_contract_tx_id, 0, 0, l_block_timestamp,
                                                         null, l_item.fund - l_item.owner_reward, 0, null, null);
            call update_sorted_oracle_voting_contracts(p_block_height, l_contract_tx_id, null, SOVC_STATE_COMPLETED,
                                                       l_tx_id, null, null);
            call update_sorted_oracle_voting_contract_committees(p_block_height, l_contract_tx_id, null, null,
                                                                 SOVC_STATE_COMPLETED, l_tx_id, null);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_oracle_voting_contract_call_prolongations(p_block_height bigint,
                                                                           p_items jsonb[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                 jsonb;
    l_tx_id                bigint;
    l_contract_address_id  bigint;
    l_contract_tx_id       bigint;
    l_start_block_height   bigint;
    l_epoch                bigint;
    l_epoch_without_growth smallint;
    l_prolong_vote_count   bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to"
            INTO l_tx_id, l_contract_address_id
            FROM transactions
            WHERE lower(hash) = lower((l_item ->> 'txHash')::text);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            l_start_block_height = (l_item ->> 'startBlockHeight')::bigint;
            l_epoch = (l_item ->> 'epoch')::bigint;
            l_epoch_without_growth = (l_item ->> 'epochWithoutGrowth')::smallint;
            l_prolong_vote_count = (l_item ->> 'prolongVoteCount')::bigint;

            INSERT INTO oracle_voting_contract_call_prolongations (call_tx_id, ov_contract_tx_id, epoch, start_block,
                                                                   vrf_seed, epoch_without_growth, prolong_vote_count)
            VALUES (l_tx_id, l_contract_tx_id, l_epoch, l_start_block_height,
                    decode(l_item ->> 'vrfSeed', 'hex'), l_epoch_without_growth, l_prolong_vote_count);

            call apply_prolongation_on_sorted_contracts(p_block_height, l_tx_id, l_contract_tx_id, l_start_block_height,
                                                        l_epoch);

            call save_oracle_voting_committee(p_block_height, l_contract_tx_id, l_item -> 'committee');

            if l_epoch_without_growth IS NOT NULL then
                call update_oracle_voting_contract_summaries(p_block_height, l_contract_tx_id, 0, 0, null, null, null,
                                                             0, null, l_epoch_without_growth);
            end if;
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_oracle_voting_contract_call_add_stakes(p_block_height bigint,
                                                                        p_items tp_oracle_voting_contract_call_add_stake[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                tp_oracle_voting_contract_call_add_stake;
    l_tx_id               bigint;
    l_amount              numeric;
    l_contract_address_id bigint;
    l_contract_tx_id      bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to", amount
            INTO l_tx_id, l_contract_address_id, l_amount
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            INSERT INTO oracle_voting_contract_call_add_stakes (call_tx_id, ov_contract_tx_id)
            VALUES (l_tx_id, l_contract_tx_id);

            call update_oracle_voting_contract_summaries(p_block_height, l_contract_tx_id, 0, 0, null, null, null,
                                                         l_amount, null, null);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_oracle_voting_contract_terminations(p_block_height bigint,
                                                                     p_items tp_oracle_voting_contract_termination[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    SOVC_STATE_TERMINATED CONSTANT smallint = 4;
    l_item                         tp_oracle_voting_contract_termination;
    l_tx_id                        bigint;
    l_contract_address_id          bigint;
    l_contract_tx_id               bigint;
    l_block_timestamp              bigint;
    l_fund                         numeric;
    l_owner_reward                 numeric;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to"
            INTO l_tx_id, l_contract_address_id
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            l_fund = null_if_negative_numeric(l_item.fund);
            l_owner_reward = null_if_negative_numeric(l_item.owner_reward);
            INSERT INTO oracle_voting_contract_terminations (termination_tx_id, ov_contract_tx_id, fund, oracle_reward,
                                                             owner_reward)
            VALUES (l_tx_id, l_contract_tx_id, l_fund,
                    null_if_negative_numeric(l_item.oracle_reward), l_owner_reward);

            SELECT "timestamp" INTO l_block_timestamp FROM blocks WHERE height = p_block_height;
            call update_oracle_voting_contract_summaries(p_block_height, l_contract_tx_id, 0, 0, null,
                                                         l_block_timestamp, l_fund - l_owner_reward, 0, null, null);
            call update_sorted_oracle_voting_contract_state(p_block_height, l_contract_tx_id, SOVC_STATE_TERMINATED,
                                                            l_tx_id);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_oracle_lock_contracts(p_block_height bigint,
                                                       p_items tp_oracle_lock_contract[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    CONTRACT_TYPE_ORACLE_LOCK CONSTANT smallint = 3;
    l_item                             tp_oracle_lock_contract;
    l_contract_address_id              bigint;
    l_tx_id                            bigint;
    l_oracle_voting_address_id         bigint;
    l_success_address_id               bigint;
    l_fail_address_id                  bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            l_contract_address_id = get_address_id_or_insert(p_block_height, l_item.contract_address);

            SELECT id INTO l_tx_id FROM transactions WHERE lower(hash) = lower(l_item.tx_hash);

            INSERT INTO contracts (tx_id, contract_address_id, "type", stake)
            VALUES (l_tx_id, l_contract_address_id, CONTRACT_TYPE_ORACLE_LOCK, l_item.stake);

            l_oracle_voting_address_id = get_address_id_or_insert(p_block_height, l_item.oracle_voting_address);

            l_success_address_id = get_address_id_or_insert(p_block_height, l_item.success_address);

            l_fail_address_id = get_address_id_or_insert(p_block_height, l_item.fail_address);

            INSERT INTO oracle_lock_contracts (contract_tx_id, oracle_voting_address_id, value, success_address_id,
                                               fail_address_id)
            VALUES (l_tx_id, l_oracle_voting_address_id, l_item.value, l_success_address_id, l_fail_address_id);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_oracle_lock_contract_call_check_oracle_votings(p_items tp_oracle_lock_contract_call_check_oracle_voting[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                tp_oracle_lock_contract_call_check_oracle_voting;
    l_tx_id               bigint;
    l_contract_address_id bigint;
    l_contract_tx_id      bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to"
            INTO l_tx_id, l_contract_address_id
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            INSERT INTO oracle_lock_contract_call_check_oracle_votings (call_tx_id, ol_contract_tx_id, oracle_voting_result)
            VALUES (l_tx_id, l_contract_tx_id, null_if_negative_bigint(l_item.oracle_voting_result)::smallint);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_oracle_lock_contract_call_pushes(p_items tp_oracle_lock_contract_call_push[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                tp_oracle_lock_contract_call_push;
    l_tx_id               bigint;
    l_contract_address_id bigint;
    l_contract_tx_id      bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to"
            INTO l_tx_id, l_contract_address_id
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            INSERT INTO oracle_lock_contract_call_pushes (call_tx_id, ol_contract_tx_id, success, oracle_voting_result,
                                                          transfer)
            VALUES (l_tx_id, l_contract_tx_id, l_item.success,
                    null_if_negative_bigint(l_item.oracle_voting_result)::smallint, l_item.transfer);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_oracle_lock_contract_terminations(p_block_height bigint,
                                                                   p_items tp_oracle_lock_contract_termination[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                tp_oracle_lock_contract_termination;
    l_tx_id               bigint;
    l_contract_address_id bigint;
    l_contract_tx_id      bigint;
    l_dest_address_id     bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to"
            INTO l_tx_id, l_contract_address_id
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            l_dest_address_id = get_address_id_or_insert(p_block_height, l_item.dest);

            INSERT INTO oracle_lock_contract_terminations (termination_tx_id, ol_contract_tx_id, dest_address_id)
            VALUES (l_tx_id, l_contract_tx_id, l_dest_address_id);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_refundable_oracle_lock_contracts(p_block_height bigint,
                                                                  p_items tp_refundable_oracle_lock_contract[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    CONTRACT_TYPE_REFUNDABLE_ORACLE_LOCK CONSTANT smallint = 5;
    l_item                                        tp_refundable_oracle_lock_contract;
    l_contract_address_id                         bigint;
    l_tx_id                                       bigint;
    l_oracle_voting_address_id                    bigint;
    l_success_address_id                          bigint;
    l_fail_address_id                             bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            l_contract_address_id = get_address_id_or_insert(p_block_height, l_item.contract_address);

            SELECT id INTO l_tx_id FROM transactions WHERE lower(hash) = lower(l_item.tx_hash);

            INSERT INTO contracts (tx_id, contract_address_id, "type", stake)
            VALUES (l_tx_id, l_contract_address_id, CONTRACT_TYPE_REFUNDABLE_ORACLE_LOCK, l_item.stake);

            l_oracle_voting_address_id = get_address_id_or_insert(p_block_height, l_item.oracle_voting_address);

            if l_item.success_address is not null and l_item.success_address <> '' then
                l_success_address_id = get_address_id_or_insert(p_block_height, l_item.success_address);
            end if;

            if l_item.fail_address is not null and l_item.fail_address <> '' then
                l_fail_address_id = get_address_id_or_insert(p_block_height, l_item.fail_address);
            end if;

            INSERT INTO refundable_oracle_lock_contracts (contract_tx_id, oracle_voting_address_id, value,
                                                          success_address_id,
                                                          fail_address_id, refund_delay, deposit_deadline,
                                                          oracle_voting_fee)
            VALUES (l_tx_id, l_oracle_voting_address_id, l_item.value, l_success_address_id, l_fail_address_id,
                    l_item.refund_delay, l_item.deposit_deadline, l_item.oracle_voting_fee);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_refundable_oracle_lock_contract_call_deposits(p_items tp_refundable_oracle_lock_contract_call_deposit[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                tp_refundable_oracle_lock_contract_call_deposit;
    l_tx_id               bigint;
    l_contract_address_id bigint;
    l_contract_tx_id      bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to"
            INTO l_tx_id, l_contract_address_id
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            INSERT INTO refundable_oracle_lock_contract_call_deposits (call_tx_id, ol_contract_tx_id, own_sum, sum, fee)
            VALUES (l_tx_id, l_contract_tx_id, l_item.own_sum, l_item.sum, l_item.fee);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_refundable_oracle_lock_contract_call_pushes(p_items tp_refundable_oracle_lock_contract_call_push[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                tp_refundable_oracle_lock_contract_call_push;
    l_tx_id               bigint;
    l_contract_address_id bigint;
    l_contract_tx_id      bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to"
            INTO l_tx_id, l_contract_address_id
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            INSERT INTO refundable_oracle_lock_contract_call_pushes (call_tx_id, ol_contract_tx_id,
                                                                     oracle_voting_exists, oracle_voting_result,
                                                                     transfer, refund_block)
            VALUES (l_tx_id, l_contract_tx_id, l_item.oracle_voting_exists,
                    null_if_negative_bigint(l_item.oracle_voting_result)::smallint,
                    l_item.transfer, null_if_zero_bigint(l_item.refund_block));
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_refundable_oracle_lock_contract_call_refunds(p_items tp_refundable_oracle_lock_contract_call_refund[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                tp_refundable_oracle_lock_contract_call_refund;
    l_tx_id               bigint;
    l_contract_address_id bigint;
    l_contract_tx_id      bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to"
            INTO l_tx_id, l_contract_address_id
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            INSERT INTO refundable_oracle_lock_contract_call_refunds (call_tx_id, ol_contract_tx_id, balance, coef)
            VALUES (l_tx_id, l_contract_tx_id, l_item.balance, l_item.coef);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_refundable_oracle_lock_contract_terminations(p_block_height bigint,
                                                                              p_items tp_refundable_oracle_lock_contract_termination[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                tp_refundable_oracle_lock_contract_termination;
    l_tx_id               bigint;
    l_contract_address_id bigint;
    l_contract_tx_id      bigint;
    l_dest_address_id     bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to"
            INTO l_tx_id, l_contract_address_id
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            l_dest_address_id = get_address_id_or_insert(p_block_height, l_item.dest);

            INSERT INTO refundable_oracle_lock_contract_terminations (termination_tx_id, ol_contract_tx_id, dest_address_id)
            VALUES (l_tx_id, l_contract_tx_id, l_dest_address_id);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_time_lock_contracts(p_block_height bigint,
                                                     p_items tp_time_lock_contract[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    CONTRACT_TYPE_TIME_LOCK CONSTANT smallint = 1;
    l_item                           tp_time_lock_contract;
    l_contract_address_id            bigint;
    l_tx_id                          bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            l_contract_address_id = get_address_id_or_insert(p_block_height, l_item.contract_address);

            SELECT id INTO l_tx_id FROM transactions WHERE lower(hash) = lower(l_item.tx_hash);

            INSERT INTO contracts (tx_id, contract_address_id, "type", stake)
            VALUES (l_tx_id, l_contract_address_id, CONTRACT_TYPE_TIME_LOCK, l_item.stake);

            INSERT INTO time_lock_contracts (contract_tx_id, "timestamp")
            VALUES (l_tx_id, l_item."timestamp");
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_time_lock_contract_call_transfers(p_block_height bigint,
                                                                   p_items tp_time_lock_contract_call_transfer[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                tp_time_lock_contract_call_transfer;
    l_tx_id               bigint;
    l_contract_address_id bigint;
    l_contract_tx_id      bigint;
    l_dest_address_id     bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to"
            INTO l_tx_id, l_contract_address_id
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            l_dest_address_id = get_address_id_or_insert(p_block_height, l_item.dest);

            INSERT INTO time_lock_contract_call_transfers (call_tx_id, tl_contract_tx_id, dest_address_id, amount)
            VALUES (l_tx_id, l_contract_tx_id, l_dest_address_id, l_item.amount);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_time_lock_contract_terminations(p_block_height bigint,
                                                                 p_items tp_time_lock_contract_termination[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                tp_time_lock_contract_termination;
    l_tx_id               bigint;
    l_contract_address_id bigint;
    l_contract_tx_id      bigint;
    l_dest_address_id     bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to"
            INTO l_tx_id, l_contract_address_id
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            l_dest_address_id = get_address_id_or_insert(p_block_height, l_item.dest);

            INSERT INTO time_lock_contract_terminations (termination_tx_id, tl_contract_tx_id, dest_address_id)
            VALUES (l_tx_id, l_contract_tx_id, l_dest_address_id);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_multisig_contracts(p_block_height bigint,
                                                    p_items tp_multisig_contract[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    CONTRACT_TYPE_MULTISIG CONSTANT smallint = 4;
    l_item                          tp_multisig_contract;
    l_contract_address_id           bigint;
    l_tx_id                         bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            l_contract_address_id = get_address_id_or_insert(p_block_height, l_item.contract_address);

            SELECT id INTO l_tx_id FROM transactions WHERE lower(hash) = lower(l_item.tx_hash);

            INSERT INTO contracts (tx_id, contract_address_id, "type", stake)
            VALUES (l_tx_id, l_contract_address_id, CONTRACT_TYPE_MULTISIG, l_item.stake);

            INSERT INTO multisig_contracts (contract_tx_id, min_votes, max_votes, state)
            VALUES (l_tx_id, l_item.min_votes, l_item.max_votes, l_item.state);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_multisig_contract_call_adds(p_block_height bigint,
                                                             p_items tp_multisig_contract_call_add[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                tp_multisig_contract_call_add;
    l_tx_id               bigint;
    l_contract_address_id bigint;
    l_contract_tx_id      bigint;
    l_address_id          bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to"
            INTO l_tx_id, l_contract_address_id
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            l_address_id = get_address_id_or_insert(p_block_height, l_item.address);

            INSERT INTO multisig_contract_call_adds (call_tx_id, ms_contract_tx_id, address_id, new_state)
            VALUES (l_tx_id, l_contract_tx_id, l_address_id, null_if_negative_bigint(l_item.new_state)::smallint);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_multisig_contract_call_sends(p_block_height bigint,
                                                              p_items tp_multisig_contract_call_send[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                tp_multisig_contract_call_send;
    l_tx_id               bigint;
    l_contract_address_id bigint;
    l_contract_tx_id      bigint;
    l_dest_address_id     bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to"
            INTO l_tx_id, l_contract_address_id
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            l_dest_address_id = get_address_id_or_insert(p_block_height, l_item.dest);

            INSERT INTO multisig_contract_call_sends (call_tx_id, ms_contract_tx_id, dest_address_id, amount)
            VALUES (l_tx_id, l_contract_tx_id, l_dest_address_id, l_item.amount);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_multisig_contract_call_pushes(p_block_height bigint,
                                                               p_items tp_multisig_contract_call_push[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                tp_multisig_contract_call_push;
    l_tx_id               bigint;
    l_contract_address_id bigint;
    l_contract_tx_id      bigint;
    l_dest_address_id     bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to"
            INTO l_tx_id, l_contract_address_id
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            l_dest_address_id = get_address_id_or_insert(p_block_height, l_item.dest);

            INSERT INTO multisig_contract_call_pushes (call_tx_id, ms_contract_tx_id, dest_address_id, amount,
                                                       vote_address_cnt, vote_amount_cnt)
            VALUES (l_tx_id, l_contract_tx_id, l_dest_address_id, l_item.amount, l_item.vote_address_cnt,
                    l_item.vote_amount_cnt);
        end loop;
END
$$;

CREATE OR REPLACE PROCEDURE save_multisig_contract_terminations(p_block_height bigint,
                                                                p_items tp_multisig_contract_termination[])
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_item                tp_multisig_contract_termination;
    l_tx_id               bigint;
    l_contract_address_id bigint;
    l_contract_tx_id      bigint;
    l_dest_address_id     bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id, "to"
            INTO l_tx_id, l_contract_address_id
            FROM transactions
            WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT tx_id
            INTO l_contract_tx_id
            FROM contracts
            WHERE contract_address_id = l_contract_address_id;

            l_dest_address_id = get_address_id_or_insert(p_block_height, l_item.dest);

            INSERT INTO multisig_contract_terminations (termination_tx_id, ms_contract_tx_id, dest_address_id)
            VALUES (l_tx_id, l_contract_tx_id, l_dest_address_id);
        end loop;
END
$$;