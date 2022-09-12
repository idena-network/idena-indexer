DO
$$
    DECLARE
        l_record record;
    BEGIN
        DROP TABLE oracle_voting_contract_authors_and_open_voters;
        CREATE TABLE oracle_voting_contract_authors_and_voters
        (
            deploy_or_vote_tx_id bigint NOT NULL,
            address_id           bigint NOT NULL,
            contract_tx_id       bigint NOT NULL
        );
        CREATE UNIQUE INDEX IF NOT EXISTS ovc_authors_and_voters_pkey ON oracle_voting_contract_authors_and_voters (deploy_or_vote_tx_id);
        CREATE UNIQUE INDEX IF NOT EXISTS ovc_authors_and_voters_ukey ON oracle_voting_contract_authors_and_voters (address_id, contract_tx_id);
        CREATE INDEX IF NOT EXISTS ovc_authors_and_voters_api ON oracle_voting_contract_authors_and_voters (address_id, contract_tx_id, deploy_or_vote_tx_id desc);

        for l_record in SELECT ovc.contract_tx_id, t.from
                        FROM oracle_voting_contracts ovc
                                 LEFT JOIN transactions t ON t.id = ovc.contract_tx_id
            loop
                INSERT INTO oracle_voting_contract_authors_and_voters
                VALUES (l_record.contract_tx_id, l_record."from", l_record.contract_tx_id);
            end loop;

        for l_record in SELECT ovccv.call_tx_id, ovccv.ov_contract_tx_id, t.from
                        FROM oracle_voting_contract_call_vote_proofs ovccv
                                 LEFT JOIN transactions t ON t.id = ovccv.call_tx_id
            loop
                INSERT INTO oracle_voting_contract_authors_and_voters
                VALUES (l_record.call_tx_id, l_record."from", l_record.ov_contract_tx_id)
                ON CONFLICT DO NOTHING;
            end loop;
    END
$$