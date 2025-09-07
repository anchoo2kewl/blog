## Deployment & Operations

This document describes how to provision dependencies, run database migrations, seed data, and run the Blog app in different environments.

### Requirements

- Go 1.25+
- Docker (for Postgres or containerized runs)
- Postgres 14+
- golang-migrate (for DB migrations)

### Environment

Copy `.env.sample` to `.env` and edit values as needed:

```
PG_USER=blog
PG_PASSWORD=1234
PG_DB=blog
PG_HOST=127.0.0.1
PG_PORT=5433
API_TOKEN=secretToken
APP_DISABLE_SIGNUP=true
```

Export into your shell for local runs:

```
export $(cat .env | xargs)
```

### Postgres via Docker

```
docker volume create postgres-volume
docker run --name pg \
  -e POSTGRES_PASSWORD=$PG_PASSWORD \
  -e POSTGRES_USER=$PG_USER \
  -p $PG_PORT:5432 \
  -v postgres-volume:/var/lib/postgresql/data \
  -d postgres
```

Create DB if needed:

```
psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_USER\?sslmode=disable -c "create database $PG_DB"
```

### Migrations

Install migrate (macOS):

```
brew install golang-migrate
```

Run migrations:

```
migrate -source file://migrations \
  -database postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB\?sslmode=disable up
```

### Seeding

Use the helper script to insert roles, users, and sample posts:

```
./scripts/server db-seed
```

### Running locally

Start DB and app together:

```
./scripts/server start --force
```

Development (live reload via Air):

```
./scripts/server dev --force
```

Rebuild + restart (ensures embedded templates refresh):

```
./scripts/server restart-blog
```

### Docker image

Build (arm64):

```
docker buildx build --platform=linux/arm64/v8 -t biswas/blog:v0.1 .
```

Build (amd64):

```
docker buildx build --platform=linux/amd64 -t biswas/blog:v0.1 .
```

Run:

```
docker run -d -p 22222:22222 --env-file .env --name blog biswas/blog:v0.1
```

Alternate port:

```
docker run -d -p 8080:8080 --env-file .env --name blog biswas/blog:v0.1 ./main --listen-addr :8080
```

### Useful script commands

```
./scripts/server start|stop|restart|status
./scripts/server start-blog|restart-blog|start-db
./scripts/server db "SELECT * FROM users LIMIT 5;"
./scripts/server db-migrate|db-seed|db-users|db-posts
./scripts/server test-go|test-playwright|test-all
./scripts/server logs blog|db
```

