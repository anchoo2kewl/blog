# My Blog

This code is aimed at replicating a simple blog. The design of this blog will be found on my [blog](https://anshumanbiswas.com).

To run using docker:

```
# For Mx Macs:
docker buildx build --progress=plain --platform=linux/arm64/v8 -t biswas/blog:v0.1 .

#For Linux on X86_64
docker buildx build --progress=plain --platform=linux/amd64 -t biswas/blog:v0.1 .

docker run -d -p 22222:22222 --name blog -v $(PWD):/go/src/blog biswas/blog:v0.1
```

or to run on port 8080:

```
docker run -d -p 8080:8080 --name blog -v $(PWD):/go/src/blog biswas/blog:v0.1 ./main --listen-addr :8080
```

To run locally for live reloading, install air:

```
curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

~/go/bin/air -c .air.toml
```
Which runs the code on 22222, unless .air.toml file has been modified.

or run without live-reloading:
```
go run main.go --listen-addr :22222
```
Now, run a Postgres Server:

```
docker pull postgres
docker volume create postgres-volume


export PG_PASSWORD=1234

docker run --name postgres -e POSTGRES_PASSWORD=$PG_PASSWORD -e POSTGRES_USER=blog -p 5433:5432 -v postgres-volume:/var/lib/postgresql/data -d postgres
```

Other tools can be used, but for mac, installing a local PSQL:
```
brew install libpq
# Add to Path, and test:
psql postgresql://blog:$PG_PASSWORD@127.0.0.1:5433/blog\?sslmode=disable
```

Install Migrate:

```
brew install golang-migrate
```

Then run the migration:

```
migrate -source file://migrations -database postgresql://blog:$PG_PASSWORD@127.0.0.1:5433/blog\?sslmode=disable up
```
