#!/bin/bash

# Script de benchmark: UUID vs ULID
# Compare les performances d'insertion et de sélection

set -e

DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="coloc_db"
DB_USER="coloc_user"
PGPASSWORD="coloc_password"

export PGPASSWORD

echo "========================================="
echo "  Benchmark: UUID vs ULID"
echo "========================================="
echo ""

# Setup
echo "1. Configuration des tables de test..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f benchmarks/setup_uuid_test.sql > /dev/null 2>&1
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f benchmarks/setup_ulid_test.sql > /dev/null 2>&1
echo "   ✓ Tables créées et peuplées avec 10 000 lignes chacune"
echo ""

# Benchmark INSERT UUID
echo "2. Test INSERT avec UUID..."
echo "   Exécution de 1000 transactions (10 clients)..."
pgbench -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME \
  -c 10 -j 2 -t 100 \
  -f benchmarks/test_uuid.sql \
  -n > /tmp/bench_uuid_insert.txt 2>&1

UUID_INSERT_TPS=$(grep "tps =" /tmp/bench_uuid_insert.txt | awk '{print $3}')
UUID_INSERT_LATENCY=$(grep "latency average" /tmp/bench_uuid_insert.txt | awk '{print $4}')
echo "   ✓ UUID INSERT: $UUID_INSERT_TPS tps, latence moyenne: $UUID_INSERT_LATENCY ms"
echo ""

# Benchmark INSERT ULID
echo "3. Test INSERT avec ULID..."
echo "   Exécution de 1000 transactions (10 clients)..."
pgbench -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME \
  -c 10 -j 2 -t 100 \
  -f benchmarks/test_ulid.sql \
  -n > /tmp/bench_ulid_insert.txt 2>&1

ULID_INSERT_TPS=$(grep "tps =" /tmp/bench_ulid_insert.txt | awk '{print $3}')
ULID_INSERT_LATENCY=$(grep "latency average" /tmp/bench_ulid_insert.txt | awk '{print $4}')
echo "   ✓ ULID INSERT: $ULID_INSERT_TPS tps, latence moyenne: $ULID_INSERT_LATENCY ms"
echo ""

# Benchmark SELECT UUID
echo "4. Test SELECT avec UUID (ORDER BY created_at)..."
echo "   Exécution de 1000 transactions (10 clients)..."
pgbench -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME \
  -c 10 -j 2 -t 100 \
  -f benchmarks/test_uuid_select.sql \
  -n > /tmp/bench_uuid_select.txt 2>&1

UUID_SELECT_TPS=$(grep "tps =" /tmp/bench_uuid_select.txt | awk '{print $3}')
UUID_SELECT_LATENCY=$(grep "latency average" /tmp/bench_uuid_select.txt | awk '{print $4}')
echo "   ✓ UUID SELECT: $UUID_SELECT_TPS tps, latence moyenne: $UUID_SELECT_LATENCY ms"
echo ""

# Benchmark SELECT ULID
echo "5. Test SELECT avec ULID (ORDER BY id)..."
echo "   Exécution de 1000 transactions (10 clients)..."
pgbench -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME \
  -c 10 -j 2 -t 100 \
  -f benchmarks/test_ulid_select.sql \
  -n > /tmp/bench_ulid_select.txt 2>&1

ULID_SELECT_TPS=$(grep "tps =" /tmp/bench_ulid_select.txt | awk '{print $3}')
ULID_SELECT_LATENCY=$(grep "latency average" /tmp/bench_ulid_select.txt | awk '{print $4}')
echo "   ✓ ULID SELECT: $ULID_SELECT_TPS tps, latence moyenne: $ULID_SELECT_LATENCY ms"
echo ""

# Taille des tables
echo "6. Comparaison de la taille des tables..."
UUID_SIZE=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT pg_size_pretty(pg_total_relation_size('users_uuid_test'));")
ULID_SIZE=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT pg_size_pretty(pg_total_relation_size('users_ulid_test'));")
echo "   UUID table: $UUID_SIZE"
echo "   ULID table: $ULID_SIZE"
echo ""

# Résumé
echo "========================================="
echo "  RÉSUMÉ DES RÉSULTATS"
echo "========================================="
echo ""
echo "INSERT Performance:"
echo "  UUID: $UUID_INSERT_TPS tps (latence: $UUID_INSERT_LATENCY ms)"
echo "  ULID: $ULID_INSERT_TPS tps (latence: $ULID_INSERT_LATENCY ms)"
echo ""
echo "SELECT Performance (tri chronologique):"
echo "  UUID (ORDER BY created_at): $UUID_SELECT_TPS tps (latence: $UUID_SELECT_LATENCY ms)"
echo "  ULID (ORDER BY id):         $ULID_SELECT_TPS tps (latence: $ULID_SELECT_LATENCY ms)"
echo ""
echo "Taille sur disque (10k lignes + index):"
echo "  UUID: $UUID_SIZE"
echo "  ULID: $ULID_SIZE"
echo ""

# Calcul des différences
UUID_INSERT_NUM=$(echo $UUID_INSERT_TPS | sed 's/[^0-9.]//g')
ULID_INSERT_NUM=$(echo $ULID_INSERT_TPS | sed 's/[^0-9.]//g')
UUID_SELECT_NUM=$(echo $UUID_SELECT_TPS | sed 's/[^0-9.]//g')
ULID_SELECT_NUM=$(echo $ULID_SELECT_TPS | sed 's/[^0-9.]//g')

if [ -n "$UUID_INSERT_NUM" ] && [ -n "$ULID_INSERT_NUM" ]; then
    INSERT_DIFF=$(awk "BEGIN {printf \"%.2f\", (($ULID_INSERT_NUM - $UUID_INSERT_NUM) / $UUID_INSERT_NUM) * 100}")
    echo "Différence INSERT: ${INSERT_DIFF}% (ULID vs UUID)"
fi

if [ -n "$UUID_SELECT_NUM" ] && [ -n "$ULID_SELECT_NUM" ]; then
    SELECT_DIFF=$(awk "BEGIN {printf \"%.2f\", (($ULID_SELECT_NUM - $UUID_SELECT_NUM) / $UUID_SELECT_NUM) * 100}")
    echo "Différence SELECT: ${SELECT_DIFF}% (ULID vs UUID)"
fi

echo ""
echo "========================================="
echo "Détails complets sauvegardés dans /tmp/bench_*.txt"
echo "========================================="
