CREATE OR REPLACE PROCEDURE save_vote_counting_step_result(p_data jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_uuid  text;
    l_votes jsonb;
    l_vote  jsonb;
BEGIN
    p_data = (p_data -> 'data')::jsonb;
    l_uuid = (SELECT uuid_in(md5(random()::text || clock_timestamp()::text)::cstring));
    INSERT INTO vote_counting_step_results (uuid, round, step, "timestamp", necessary_votes_count, checked_round_votes)
    VALUES (l_uuid, (p_data ->> 'round')::bigint, (p_data ->> 'step')::smallint, (p_data ->> 'timestamp')::bigint,
            (p_data ->> 'necessaryVotesCount')::integer, (p_data ->> 'checkedRoundVotes')::integer);

    l_votes = (p_data -> 'votes')::jsonb;
    if l_votes is not null then
        for i in 0..jsonb_array_length(l_votes) - 1
            loop
                l_vote = (l_votes ->> i)::jsonb;
                INSERT INTO vote_counting_step_result_votes (uuid, voter, parent_hash, voted_hash, turn_offline, upgrade)
                VALUES (l_uuid, (l_vote ->> 'voter')::text, (l_vote ->> 'parentHash')::text,
                        (l_vote ->> 'votedHash')::text, (l_vote ->> 'turnOffline')::boolean,
                        (l_vote ->> 'upgrade')::integer);
            end loop;
    end if;

END
$$;

CREATE OR REPLACE PROCEDURE save_vote_counting_result(p_data jsonb)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_err             text;
    l_uuid            text;
    l_validators      jsonb;
    l_validators_size integer;
    l_addresses       jsonb;
    l_cert            jsonb;
    l_votes           jsonb;
    l_vote            jsonb;
BEGIN
    p_data = (p_data -> 'data')::jsonb;
    l_uuid = (SELECT uuid_in(md5(random()::text || clock_timestamp()::text)::cstring));
    l_validators = (p_data -> 'validators')::jsonb;
    if l_validators is not null then
        l_validators_size = (l_validators ->> 'size')::integer;
    end if;
    l_err = (p_data ->> 'err')::text;
    if l_err is not null and char_length(l_err) > 200 then
        l_err = substring(l_err from 1 for 200);
    end if;
    INSERT INTO vote_counting_results (uuid, round, step, "timestamp", hash, err, validators_size)
    VALUES (l_uuid, (p_data ->> 'round')::bigint, (p_data ->> 'step')::smallint, (p_data ->> 'timestamp')::bigint,
            (p_data ->> 'hash')::text, l_err, l_validators_size);

    if l_validators is not null then
        l_addresses = (l_validators -> 'original')::jsonb;
        if l_addresses is not null then
            for i in 0..jsonb_array_length(l_addresses) - 1
                loop
                    INSERT INTO vote_counting_result_validators_original (uuid, address)
                    VALUES (l_uuid, (l_addresses ->> i)::text);
                end loop;
        end if;
        l_addresses = (l_validators -> 'validators')::jsonb;
        if l_addresses is not null then
            for i in 0..jsonb_array_length(l_addresses) - 1
                loop
                    INSERT INTO vote_counting_result_validators_validators (uuid, address)
                    VALUES (l_uuid, (l_addresses ->> i)::text);
                end loop;
        end if;
        l_addresses = (l_validators -> 'approvedValidators')::jsonb;
        if l_addresses is not null then
            for i in 0..jsonb_array_length(l_addresses) - 1
                loop
                    INSERT INTO vote_counting_result_validators_approved_validators (uuid, address)
                    VALUES (l_uuid, (l_addresses ->> i)::text);
                end loop;
        end if;
    end if;

    l_cert = (p_data -> 'cert')::jsonb;
    if l_cert is not null then
        l_votes = (l_cert -> 'votes')::jsonb;
        if l_votes is not null then
            for i in 0..jsonb_array_length(l_votes) - 1
                loop
                    l_vote = (l_votes ->> i)::jsonb;
                    INSERT INTO vote_counting_result_cert_votes (uuid, voter, parent_hash, voted_hash, turn_offline, upgrade)
                    VALUES (l_uuid, (l_vote ->> 'voter')::text, (l_vote ->> 'parentHash')::text,
                            (l_vote ->> 'votedHash')::text, (l_vote ->> 'turnOffline')::boolean,
                            (l_vote ->> 'upgrade')::integer);
                end loop;
        end if;
    end if;
END
$$;

CREATE OR REPLACE PROCEDURE save_proof_proposal(p_data jsonb)
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    p_data = (p_data -> 'data')::jsonb;
    INSERT INTO proof_proposals (round, "timestamp", proposer, "hash", modifier, vrf_score)
    VALUES ((p_data ->> 'round')::bigint, (p_data ->> 'timestamp')::bigint, (p_data ->> 'proposer')::text,
            (p_data ->> 'hash')::text, (p_data ->> 'modifier')::integer, (p_data ->> 'vrfScore')::double precision);
END
$$;

CREATE OR REPLACE PROCEDURE save_block_proposal(p_data jsonb)
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    p_data = (p_data -> 'data')::jsonb;
    INSERT INTO block_proposals (height, receiving_time, proposer, "hash")
    VALUES ((p_data ->> 'height')::bigint, (p_data ->> 'receivingTime')::bigint, (p_data ->> 'proposer')::text,
            (p_data ->> 'hash')::text);
END
$$;

CREATE OR REPLACE PROCEDURE delete_old_vote_counting_data(p_changes_blocks_count smallint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    l_max_height bigint;
BEGIN
    SELECT max(round) INTO l_max_height FROM vote_counting_results;
    if l_max_height is null then
        return;
    end if;
    if NOT (SELECT exists(SELECT 1
                          FROM vote_counting_results
                          WHERE round <= l_max_height - p_changes_blocks_count)) then
        return;
    end if;

    DROP TABLE IF EXISTS vote_counting_step_results_old;
    DROP TABLE IF EXISTS vote_counting_step_result_votes_old;
    DROP TABLE IF EXISTS vote_counting_results_old;
    DROP TABLE IF EXISTS vote_counting_result_validators_original_old;
    DROP TABLE IF EXISTS vote_counting_result_validators_validators_old;
    DROP TABLE IF EXISTS vote_counting_result_validators_approved_validators_old;
    DROP TABLE IF EXISTS vote_counting_result_cert_votes_old;
    DROP TABLE IF EXISTS proof_proposals_old;
    DROP TABLE IF EXISTS block_proposals_old;

    ALTER TABLE vote_counting_step_results
        RENAME TO vote_counting_step_results_old;
    ALTER TABLE vote_counting_step_results_old
        RENAME CONSTRAINT vote_counting_step_results_pkey TO vote_counting_step_results_old_pkey;
    ALTER INDEX vote_counting_step_results_unique_idx1 RENAME TO vote_counting_step_results_old_unique_idx1;
    ALTER INDEX vote_counting_step_results_idx1 RENAME TO vote_counting_step_results_old_idx1;

    ALTER TABLE vote_counting_step_result_votes
        RENAME TO vote_counting_step_result_votes_old;
    ALTER INDEX vote_counting_step_result_votes_idx1 RENAME TO vote_counting_step_result_votes_old_idx1;

    ALTER TABLE vote_counting_results
        RENAME TO vote_counting_results_old;
    ALTER INDEX vote_counting_results_unique_idx1 RENAME TO vote_counting_results_old_unique_idx1;
    ALTER INDEX vote_counting_results_idx1 RENAME TO vote_counting_results_old_idx1;

    ALTER TABLE vote_counting_result_validators_original
        RENAME TO vote_counting_result_validators_original_old;
    ALTER INDEX vote_counting_result_validators_original_idx1 RENAME TO vote_counting_result_validators_original_old_idx1;

    ALTER TABLE vote_counting_result_validators_validators
        RENAME TO vote_counting_result_validators_validators_old;
    ALTER INDEX vote_counting_result_validators_validators_idx1 RENAME TO vote_counting_result_validators_validators_old_idx1;

    ALTER TABLE vote_counting_result_validators_approved_validators
        RENAME TO vote_counting_result_validators_approved_validators_old;
    ALTER INDEX vote_counting_result_validators_approved_validators_idx1 RENAME TO vote_counting_result_validators_approved_validators_old_idx1;

    ALTER TABLE vote_counting_result_cert_votes
        RENAME TO vote_counting_result_cert_votes_old;
    ALTER INDEX vote_counting_result_cert_votes_idx1 RENAME TO vote_counting_result_cert_votes_old_idx1;

    ALTER TABLE proof_proposals
        RENAME TO proof_proposals_old;
    ALTER INDEX proof_proposals_idx1 RENAME TO proof_proposals_old_idx1;

    ALTER TABLE block_proposals
        RENAME TO block_proposals_old;
    ALTER INDEX block_proposals_idx1 RENAME TO block_proposals_old_idx1;

    call create_vote_counting_tables();
END
$$;