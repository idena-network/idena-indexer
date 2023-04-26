ALTER TABLE contract_verifications
    ADD COLUMN file_name character varying(50);
ALTER TABLE contract_verifications
    ADD COLUMN error_message character varying(200);
DROP PROCEDURE update_contract_verification_state;