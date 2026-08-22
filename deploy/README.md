# Production Compose overlay

The production overlay adds rolling updates, bounded restart policies, and log rotation while preserving the health checks and read-only application filesystem from the base Compose file.

Validate the merged configuration before deployment:

```sh
docker compose -f compose.yaml -f deploy/compose.production.yaml config --quiet
```

Deploy with environment-specific database credentials and allowed origin values supplied by the runtime:

```sh
docker compose -f compose.yaml -f deploy/compose.production.yaml up --build -d
```
