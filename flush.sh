#!/bin/bash

# Down containers
docker compose down
docker compose -f proto/gatewaybuf/docker-compose.yaml down
docker compose -f proto/reportbuf/docker-compose.yaml down
docker compose -f gateway/docker-compose.yaml down
docker compose -f report/docker-compose.yaml down
docker compose -f producer/docker-compose.yaml down

rm -rf volumes
rm -rf proto/gatewaybuf/git.infra.egt.com
rm -rf proto/reportbuf/git.infra.egt.com
