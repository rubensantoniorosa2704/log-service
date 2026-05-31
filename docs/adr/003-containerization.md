# ADR-003: Containerization Strategy

**Status:** Accepted  
**Date:** 2026-05-30  
**Author:** log-service team

---

## Context

log-service requires a reproducible, environment-independent way to run both the application and its MongoDB dependency. Without containerization, developers must install and configure MongoDB locally, manage Go toolchain versions manually, and ensure environment variables are set correctly — all of which introduce friction and inconsistency across machines.

Docker was chosen as the containerization platform. This ADR documents the decisions made in the `Dockerfile` and `docker-compose.yml`, the rationale behind each choice, and the standard commands for operating the service.

---

## Decision

### Multi-stage Dockerfile with `scratch` as the final base

The `Dockerfile` uses a two-stage build:

1. **Builder stage** (`golang:1.25.3-alpine`): compiles the binary with `CGO_ENABLED=0` to produce a fully static binary.
2. **Final stage** (`scratch`): an empty base image containing only the compiled binary and the `docs/` directory.

Using `scratch` instead of `alpine` or `debian` eliminates the OS layer entirely. Since the binary is statically linked, no runtime dependencies exist. The resulting image is minimal (~10MB vs ~20MB with alpine), has no shell, no package manager, and no OS utilities — reducing the attack surface to the application binary itself.

The `docs/` directory is copied into the final image to serve the Swagger UI at runtime.

> **Note:** If the service ever needs to make outbound HTTPS requests (e.g., to an external API), TLS certificate bundles must be added explicitly:
> ```dockerfile
> COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
> ```
> This is not required for the current scope, where the only external connection is to MongoDB over the internal Docker network.

### Docker Compose for local development

`docker-compose.yml` defines two services:

- **`mongodb`**: MongoDB 6.0 with a named volume for data persistence across restarts. A healthcheck using `mongosh` ensures the app container only starts after MongoDB is ready to accept connections.
- **`app`**: built from the local `Dockerfile`. Waits for `mongodb` to pass its healthcheck before starting (`depends_on: condition: service_healthy`).

The `app` service reads all configuration from environment variables supplied via the `.env` file. No configuration is hardcoded in the Compose file.

### Environment variable configuration

All runtime configuration is injected via environment variables. The `.env` file (not committed to version control) is the single source of configuration for local development.

| Variable | Description | Example value |
|---|---|---|
| `MONGO_URI` | Full MongoDB connection string | `mongodb://root:rootpassword@mongodb:27017/loggingdb?authSource=admin` |
| `MONGO_DB_NAME` | Database name | `loggingdb` |
| `PORT` | Port the application listens on inside the container | `8080` |
| `APP_PORT` | Host port mapped to the container's `8080` | `8080` |

> **Important:** `MONGO_URI` must use the Docker Compose service name (`mongodb`) as the hostname, not `localhost`. Within the Docker network, containers resolve each other by service name. Using `localhost` will cause the app to fail to connect to MongoDB.

The `.env.example` file documents all required variables and serves as the template for new environments. Copy it to `.env` before running the stack:

```bash
cp .env.example .env
```

---

## Standard Commands

### Start the full stack

```bash
docker-compose up -d
```

Starts MongoDB and the application in detached mode. On first run, Docker builds the application image automatically.

### Rebuild the application image

Required after any code change:

```bash
docker-compose up -d --build
```

### View application logs

```bash
docker-compose logs -f app
```

### Stop the stack

```bash
docker-compose down
```

Data in the `mongo-data` volume is preserved. To also remove the volume (full reset):

```bash
docker-compose down -v
```

### Run the application only (MongoDB must be running separately)

```bash
docker build -t log-service .
docker run -p 8080:8080 --env-file .env log-service
```

---

## Alternatives Considered

### `alpine` as the final base image

`alpine` (~7MB) is a common choice for Go services because it includes a shell and `apk`, which simplifies debugging. It was used in the original `Dockerfile` to provide `curl` for the healthcheck.

**Rejected in favor of `scratch`:** the healthcheck was removed (the `/` route does not exist in this API, making the check unreliable), and the debugging convenience of a shell does not justify the larger attack surface in a service that handles application log data. If runtime debugging is needed, `docker-compose exec` can be replaced with a temporary `alpine`-based image.

### Hardcoding credentials in `docker-compose.yml`

The MongoDB credentials (`root`/`rootpassword`) are defined in `docker-compose.yml` for the `mongodb` service initialization. These are development-only credentials. For any environment beyond local development, credentials must be supplied via secrets management (e.g., Docker Secrets, AWS Secrets Manager) and must not be committed to version control.

---

## Consequences

- The application image is minimal and contains no OS utilities, reducing the attack surface.
- All configuration is externalized to `.env`, making the stack portable across developer machines.
- The `depends_on: condition: service_healthy` ensures the app never starts before MongoDB is ready, eliminating a class of startup race condition errors.
- The `.env` file is excluded from version control via `.gitignore`. New contributors must create it from `.env.example` before running the stack.
- Rebuilding the image after code changes requires an explicit `--build` flag; `docker-compose up -d` alone will not pick up source changes.

---

## References

- [Docker multi-stage builds](https://docs.docker.com/build/building/multi-stage/)
- [Docker Compose `depends_on` with healthcheck](https://docs.docker.com/compose/how-tos/startup-order/)
- [Go static binaries and `scratch`](https://pkg.go.dev/cmd/go#hdr-Build_constraints)
- [MongoDB Docker image](https://hub.docker.com/_/mongo)
