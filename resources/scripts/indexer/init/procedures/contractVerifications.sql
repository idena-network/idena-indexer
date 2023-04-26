CREATE OR REPLACE FUNCTION save_contract_pending_verification(p_contract_address text,
                                                              p_timestamp bigint,
                                                              p_data bytea)
    RETURNS text
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    CONTRACT_TYPE_CONTRACT      CONSTANT smallint = 6;
    VERIFICATION_STATE_PENDING  constant smallint = 0;
    VERIFICATION_STATE_VERIFIED constant smallint = 1;
    VERIFICATION_STATE_FAILED   constant smallint = 2;
    l_contract_address_id                bigint;
    l_state                              smallint;
BEGIN
    SELECT c.contract_address_id, cv.state
    INTO l_contract_address_id, l_state
    FROM contracts c
             LEFT JOIN contract_verifications cv ON cv.contract_address_id = c.contract_address_id
    WHERE c.contract_address_id = (SELECT id FROM addresses WHERE lower(address) = lower(p_contract_address))
      AND c.type = CONTRACT_TYPE_CONTRACT;

    if l_contract_address_id is null then
        return 'not_wasm_contract';
    end if;

    if l_state = VERIFICATION_STATE_PENDING then
        return 'already_submitted';
    end if;

    if l_state = VERIFICATION_STATE_VERIFIED then
        return 'already_verified';
    end if;

    if l_state = VERIFICATION_STATE_FAILED then
        DELETE FROM contract_verifications WHERE contract_address_id = l_contract_address_id;
    end if;

    INSERT INTO contract_verifications (contract_address_id, state, state_timestamp, "data")
    VALUES (l_contract_address_id, VERIFICATION_STATE_PENDING, p_timestamp, p_data);
    return null;
END
$$;

CREATE OR REPLACE PROCEDURE update_contract_verification_state(p_contract_address text,
                                                               p_state smallint,
                                                               p_timestamp bigint,
                                                               p_data bytea,
                                                               p_error_message text)
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    UPDATE contract_verifications
    SET state           = p_state,
        state_timestamp = p_timestamp,
        "data"          = p_data,
        error_message   = p_error_message
    WHERE contract_address_id = (SELECT id FROM addresses WHERE lower(address) = lower(p_contract_address));
END
$$;