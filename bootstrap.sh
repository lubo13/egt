#!/bin/bash

# Proto buf building
docker compose -f proto/gatewaybuf/docker-compose.yaml up -d buf_build_gateway_proto
docker compose -f proto/reportbuf/docker-compose.yaml up -d buf_build_report_proto
sleep 5s
docker compose -f proto/gatewaybuf/docker-compose.yaml up -d module_init_gateway_proto
docker compose -f proto/reportbuf/docker-compose.yaml up -d module_init_report_proto

echo "ProtoBuf definitions were built..."

# Run infrastructure services
docker compose up -d

echo "Sleep 15 seconds because of database..."

sleep 15s

docker exec kafkabroker1 kafka-topics --bootstrap-server kafkabroker1:9092 --create --if-not-exists --topic device-events --partitions 3 --replication-factor 1

echo "Infrastructure services are running..."

# Prepare Gateway service database
docker compose -f gateway/docker-compose.yaml up -d gateway_migrator

# Prepare Report service database
docker compose -f report/docker-compose.yaml up -d report_migrator

# Run Gateway service http server and grpc server
docker compose -f gateway/docker-compose.yaml up -d gateway_server

# Run Producer service worker
docker compose -f producer/docker-compose.yaml up -d producer

# Run Report service kafka consumer
docker compose -f report/docker-compose.yaml up -d report_consumer

# Run Report service grpc server
docker compose -f report/docker-compose.yaml up -d report_grpc_server
