-- Script pgbench: Test SELECT avec ULID (tri par id qui contient le timestamp)
SELECT id, email, nom, prenom FROM users_ulid_test ORDER BY id DESC LIMIT 100;
