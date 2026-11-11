# Deployment

Runs the calculator service, the web app and Caddy together with Docker Compose. Caddy is
the only container reachable from outside. It sends `/api` requests to the service and
everything else to the web app.

```
browser ──► Caddy (ports 80, 443) ──┬── /api/* ──► api (Go, 8081)
                                    └── rest   ──► web (Node, 3000)
```

## Run

```sh
cd deploy
cp .env.example .env
docker compose up -d --build
```

Open http://localhost. To follow the service's logs: `docker compose logs -f api`. To stop:
`docker compose down`.

Settings in `.env`:

- `SITE_ADDRESS`: the address Caddy serves. Default `:80`. Set it to a domain for HTTPS.
- `LOG_LEVEL`: passed to the calculator service. Default `info`.
- `OTEL_EXPORTER_OTLP_ENDPOINT`: passed to the service. Set it to your collector (Jaeger,
  Grafana Alloy, an OpenTelemetry Collector) to export traces and metrics.

## HTTPS

1. Point the domain's DNS at the server.
2. Open ports 80 and 443.
3. Set `SITE_ADDRESS=calc.example.com`.

Caddy gets and renews the certificate on its own.

## Notes

- Each image is built from its own folder, `../backend` and `../frontend`.
- Containers have no Docker health checks; a crashed one is restarted. An orchestrator or
  load balancer can probe the service at `GET /healthz`.
- The service keeps no data, so several copies can run: `docker compose up --scale api=3`.
- No container runs as root.

To serve the web app from a different domain than the API, build it with
`--build-arg VITE_API_URL=https://api.example.com` and set `ALLOWED_ORIGINS` on the service
to the web app's address.
