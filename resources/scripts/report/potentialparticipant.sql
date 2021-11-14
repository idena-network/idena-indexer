CREATE MATERIALIZED VIEW report.mv_potentialparticipant AS
SELECT total.epoch,
       total.candidate_v,
       part.candidate_p,
       total.candidate_t,
       total.newbie_v,
       part.newbie_p,
       total.newbie_t,
       total.verified_v,
       part.verified_p,
       total.verified_t,
       total.human_v,
       part.human_p,
       total.human_t,
       total.suspended_v,
       part.suspended_p,
       total.suspended_t,
       total.zombie_v,
       part.zombie_p,
       total.zombie_t
FROM (SELECT ei.epoch + 1 AS epoch,
             sum(CASE
                     WHEN prevs.state = 2 THEN 1
                     ELSE 0
                 END)     AS candidate_t,
             sum(CASE
                     WHEN prevs.state = 2 AND (s.state = ANY (ARRAY [3, 7, 8])) THEN 1
                     ELSE 0
                 END)     AS candidate_v,
             sum(CASE
                     WHEN prevs.state = 7 THEN 1
                     ELSE 0
                 END)     AS newbie_t,
             sum(CASE
                     WHEN prevs.state = 7 AND (s.state = ANY (ARRAY [3, 7, 8])) THEN 1
                     ELSE 0
                 END)     AS newbie_v,
             sum(CASE
                     WHEN prevs.state = 3 THEN 1
                     ELSE 0
                 END)     AS verified_t,
             sum(CASE
                     WHEN prevs.state = 3 AND (s.state = ANY (ARRAY [3, 7, 8])) THEN 1
                     ELSE 0
                 END)     AS verified_v,
             sum(CASE
                     WHEN prevs.state = 8 THEN 1
                     ELSE 0
                 END)     AS human_t,
             sum(CASE
                     WHEN prevs.state = 8 AND (s.state = ANY (ARRAY [3, 7, 8])) THEN 1
                     ELSE 0
                 END)     AS human_v,
             sum(CASE
                     WHEN prevs.state = 4 THEN 1
                     ELSE 0
                 END)     AS suspended_t,
             sum(CASE
                     WHEN prevs.state = 4 AND (s.state = ANY (ARRAY [3, 7, 8])) THEN 1
                     ELSE 0
                 END)     AS suspended_v,
             sum(CASE
                     WHEN prevs.state = 6 THEN 1
                     ELSE 0
                 END)     AS zombie_t,
             sum(CASE
                     WHEN prevs.state = 6 AND (s.state = ANY (ARRAY [3, 7, 8])) THEN 1
                     ELSE 0
                 END)     AS zombie_v
      FROM indexer.epoch_identities ei
               JOIN indexer.address_states s ON s.id = ei.address_state_id
               JOIN indexer.address_states prevs ON prevs.id = s.prev_id
      WHERE prevs.state = ANY (ARRAY [2, 3, 4, 6, 7, 8])
      GROUP BY ei.epoch) total
         JOIN (SELECT p.epoch,
                      sum(CASE
                              WHEN prevs.state = 2 THEN 1
                              ELSE 0
                          END) AS candidate_p,
                      sum(CASE
                              WHEN prevs.state = 7 THEN 1
                              ELSE 0
                          END) AS newbie_p,
                      sum(CASE
                              WHEN prevs.state = 3 THEN 1
                              ELSE 0
                          END) AS verified_p,
                      sum(CASE
                              WHEN prevs.state = 8 THEN 1
                              ELSE 0
                          END) AS human_p,
                      sum(CASE
                              WHEN prevs.state = 4 THEN 1
                              ELSE 0
                          END) AS suspended_p,
                      sum(CASE
                              WHEN prevs.state = 6 THEN 1
                              ELSE 0
                          END) AS zombie_p
               FROM report.mv_participants p
                        JOIN indexer.epoch_identities ei ON ei.epoch + 1 = p.epoch AND ei.address_id = p.address_id
                        JOIN indexer.address_states s ON s.id = ei.address_state_id
                        JOIN indexer.address_states prevs ON prevs.id = s.prev_id
               WHERE prevs.state = ANY (ARRAY [2, 3, 4, 6, 7, 8])
               GROUP BY p.epoch) part ON part.epoch = total.epoch
ORDER BY total.epoch;

CREATE OR REPLACE PROCEDURE report.refresh_pot_participant()
    LANGUAGE 'plpgsql'
AS
$$
BEGIN
    call report.refresh_participants();
    refresh materialized view report.mv_potentialparticipant;
END
$$;

INSERT INTO report.dynamic_endpoints
("name", refresh_procedure, refresh_period, refresh_delay_minutes, endpoint_method, "limit")
VALUES ('report.mv_potentialparticipant', 'report.refresh_pot_participant', 'e', 30, 'PotentialParticipants',
        null);