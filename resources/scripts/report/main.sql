CREATE SCHEMA report;
CREATE TABLE report.dynamic_endpoints
(
    name                  character varying(30) NOT NULL,
    refresh_procedure     character varying(30),
    refresh_period        character varying(1),
    refresh_delay_minutes smallint,
    endpoint_method       character varying(30),
    "limit"               smallint
);

CREATE TABLE report.dynamic_endpoint_states
(
    name              character varying(100) NOT NULL,
    refresh_time      bigint,
    refresh_epoch     smallint,
    last_refresh_time bigint,
    CONSTRAINT data_state_pkey PRIMARY KEY (name)
);