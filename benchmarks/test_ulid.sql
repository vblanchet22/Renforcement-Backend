-- Script pgbench: Test INSERT avec ULID
\set random_id random(1, 1000000)
INSERT INTO users_ulid_test (email, nom, prenom, telephone)
VALUES ('bench' || :random_id || '@test.com', 'BenchNom' || :random_id, 'BenchPrenom' || :random_id, '0612345678')
ON CONFLICT (email) DO NOTHING;
