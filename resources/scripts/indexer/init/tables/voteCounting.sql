CREATE OR REPLACE PROCEDURE create_vote_counting_tables()
    LANGUAGE 'plpgsql'
AS
$$
BEGIN

    CREATE TABLE IF NOT EXISTS vote_counting_step_results
    (
        uuid                  character(36) NOT NULL,
        round                 bigint        NOT NULL,
        step                  smallint,
        timestamp             bigint,
        necessary_votes_count integer,
        checked_round_votes   integer,
        CONSTRAINT vote_counting_step_results_pkey PRIMARY KEY (uuid)
    );
    CREATE UNIQUE INDEX IF NOT EXISTS vote_counting_step_results_unique_idx1 on vote_counting_step_results (LOWER(uuid));
    CREATE INDEX IF NOT EXISTS vote_counting_step_results_idx1 on vote_counting_step_results (round);

    CREATE TABLE IF NOT EXISTS vote_counting_step_result_votes
    (
        uuid         character(36) NOT NULL,
        voter        character(42),
        parent_hash  character(66),
        voted_hash   character(66),
        turn_offline boolean,
        upgrade      integer
    );
    CREATE INDEX IF NOT EXISTS vote_counting_step_result_votes_idx1 on vote_counting_step_result_votes (uuid);

    CREATE TABLE IF NOT EXISTS vote_counting_results
    (
        uuid            character(36) NOT NULL,
        round           bigint        NOT NULL,
        step            smallint,
        timestamp       bigint,
        hash            character(66),
        err             character varying(200),
        validators_size integer
    );
    CREATE UNIQUE INDEX IF NOT EXISTS vote_counting_results_unique_idx1 on vote_counting_results (LOWER(uuid));
    CREATE INDEX IF NOT EXISTS vote_counting_results_idx1 on vote_counting_results (round);

    CREATE TABLE IF NOT EXISTS vote_counting_result_validators_original
    (
        uuid    character(36) NOT NULL,
        address character(42)
    );
    CREATE INDEX IF NOT EXISTS vote_counting_result_validators_original_idx1 on vote_counting_result_validators_original (uuid);

    CREATE TABLE IF NOT EXISTS vote_counting_result_validators_addresses
    (
        uuid    character(36) NOT NULL,
        address character(42)
    );
    CREATE INDEX IF NOT EXISTS vote_counting_result_validators_addresses_idx1 on vote_counting_result_validators_addresses (uuid);

    CREATE TABLE IF NOT EXISTS vote_counting_result_cert_votes
    (
        uuid         character(36) NOT NULL,
        voter        character(42),
        parent_hash  character(66),
        voted_hash   character(66),
        turn_offline boolean,
        upgrade      integer
    );
    CREATE INDEX IF NOT EXISTS vote_counting_result_cert_votes_idx1 on vote_counting_result_cert_votes (uuid);

    CREATE TABLE IF NOT EXISTS proof_proposals
    (
        round     bigint NOT NULL,
        timestamp bigint,
        proposer  character(42),
        hash      character(66),
        modifier  integer,
        vrf_score double precision
    );
    CREATE INDEX IF NOT EXISTS proof_proposals_idx1 on proof_proposals (round);

    CREATE TABLE IF NOT EXISTS block_proposals
    (
        height         bigint NOT NULL,
        receiving_time bigint,
        proposer       character(42),
        hash           character(66)
    );
    CREATE INDEX IF NOT EXISTS block_proposals_idx1 on block_proposals (height);

END
$$;

call create_vote_counting_tables();