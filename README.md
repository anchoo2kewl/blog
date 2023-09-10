# My Blog

This code is aimed at replicating a simple blog. The design of this blog will be found on my [blog](https://anshumanbiswas.com).

To run using docker:

```
# For Mx Macs:
docker buildx build --progress=plain --platform=linux/arm64/v8 -t biswas/blog:v0.1 .

#For Linux on X86_64
docker buildx build --progress=plain --platform=linux/amd64 -t biswas/blog:v0.1 .

docker run -d -p 22222:22222 --name blog -v $(pwd):/go/src/blog biswas/blog:v0.1
```

or to run on port 8080:

```
docker run -d -p 8080:8080 --name blog -v $(pwd):/go/src/blog biswas/blog:v0.1 ./main --listen-addr :8080
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

# Change this to a more secure password
export PG_PASSWORD=1234

export PG_PORT=5433
export PG_USER=blog
export PG_DB=blog
export PG_HOST=127.0.0.1
export APP_DISABLE_SIGNUP=true

docker run --name pg -e POSTGRES_PASSWORD=$PG_PASSWORD -e POSTGRES_USER=$PG_USER -p $PG_PORT:5432 -v postgres-volume:/var/lib/postgresql/data -d postgres
```

Other tools can be used, but for mac, installing a local PSQL:
```
brew install libpq
# Add to Path, and test:
psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_USER\?sslmode=disable
```

Install Migrate:

```
brew install golang-migrate
```

Then run the migration:

```
migrate -source file://migrations -database postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_USER\?sslmode=disable up
```

Prepare the DB:

```
psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_USER\?sslmode=disable -c "INSERT INTO ROLES (role_name) values ('Commenter')"
#Check
psql postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_USER\?sslmode=disable -c 'SELECT * FROM roles'
```