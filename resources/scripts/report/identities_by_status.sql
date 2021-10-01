CREATE MATERIALIZED VIEW report.mv_identities_by_status AS
SELECT ei.epoch + 1 AS epoch,
       sum(CASE
               WHEN s.state = 7 THEN 1
               ELSE 0
           END)     AS newbie,
       sum(CASE
               WHEN s.state = 3 THEN 1
               ELSE 0
           END)     AS verified,
       sum(CASE
               WHEN s.state = 8 THEN 1
               ELSE 0
           END)     AS human,
       sum(CASE
               WHEN s.state = 4 THEN 1
               ELSE 0
           END)     AS suspended,
       sum(CASE
               WHEN s.state = 6 THEN 1
               ELSE 0
           END)     AS zombie
FROM indexer.epoch_identities ei
         JOIN indexer.address_states s ON s.id = ei.address_state_id
GROUP BY ei.epoch
ORDER BY ei.epoch;

CREATE OR REPLACE PROCEDURE report.refresh_id_by_st()
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    refresh materialized view report.mv_identities_by_status;
END
$$;

INSERT INTO report.dynamic_endpoints
("name", refresh_procedure, refresh_period, refresh_delay_minutes, endpoint_method, "limit")
VALUES ('report.mv_identities_by_status', 'report.refresh_id_by_st', 'e', 30, 'IdentitiesByStatus', null);