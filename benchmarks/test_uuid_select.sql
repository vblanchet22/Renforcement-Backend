-- Script pgbench: Test SELECT avec UUID (tri par created_at)
SELECT id, email, nom, prenom FROM users_uuid_test ORDER BY created_at DESC LIMIT 100;
