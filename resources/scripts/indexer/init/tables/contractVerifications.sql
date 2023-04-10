CREATE TABLE IF NOT EXISTS dic_contract_verification_states
(
    id   smallint              NOT NULL,
    name character varying(20) NOT NULL,
    CONSTRAINT dic_contract_verification_states_pkey PRIMARY KEY (id),
    CONSTRAINT dic_contract_verification_states_name_key UNIQUE (name)
);

INSERT INTO dic_contract_verification_states
VALUES (0, 'Pending')
ON CONFLICT DO NOTHING;
INSERT INTO dic_contract_verification_states
VALUES (1, 'Verified')
ON CONFLICT DO NOTHING;
INSERT INTO dic_contract_verification_states
VALUES (2, 'Failed')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS contract_verifications
(
    contract_address_id bigint   NOT NULL,
    state               smallint NOT NULL,
    state_timestamp     bigint   NOT NULL,
    "data"              bytea
);
CREATE UNIQUE INDEX IF NOT EXISTS contract_verifications_pkey ON contract_verifications (contract_address_id);
CREATE INDEX IF NOT EXISTS contract_verifications_pending ON contract_verifications (contract_address_id) WHERE state = 0;
