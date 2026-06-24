# Gateway Health Check (Service A)
```bash
    grpcurl -plaintext -emit-defaults 'localhost:8899' grpc.health.v1.Health.Check
```

# Producer Health Check
```bash
    grpcurl -plaintext -emit-defaults 'localhost:8999' grpc.health.v1.Health.Check
```

# Report Consumer Health Check (Service B)
```bash
    grpcurl -plaintext -emit-defaults 'localhost:8989' grpc.health.v1.Health.Check
```

# Report Server Health Check (Service B)
```bash
    grpcurl -plaintext -emit-defaults 'localhost:8988' grpc.health.v1.Health.Check
```

# Report Server List Event (Service B)
```bash
    grpcurl -plaintext -emit-defaults 'localhost:8988' report.api.v1.DeviceEventService.ListEvent
```

# Report Server Get Event by ID (Service B)
```bash
    grpcurl \
	-plaintext \
	-H 'id':'e02d933c-832d-4e1c-a3c2-afc18e90cb07' \
	-emit-defaults \
	-d '{"id":"e02d933c-832d-4e1c-a3c2-afc18e90c111"}' \
	'localhost:8988' \
	report.api.v1.DeviceEventService.GetEvent
```
