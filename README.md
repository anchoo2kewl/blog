# My Blog

This code is aimed at replicating a simple blog. The design of this blog will be found on my [blog](https://anshumanbiswas.com).

To run using docker:

```
docker build -t biswas/blog:v0.1 .

docker run -d -p 8080:8080 --name blog -v $(PWD):/go/src/blog biswas/blog:v0.1
```

or to run on port 3000:

```
docker run -d -p 3000:3000 --name blog -v $(PWD):/go/src/blog biswas/blog:v0.1 ./main --listen-addr :3000
```

To run locally for live reloading, install air:

```
curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

~/go/bin/air -c .air.toml
```
Which runs the code on 8080, unless .air.toml file has been modified.

Now, run a Postgres Server:

```
docker pull postgres
docker volume create postgres-volume
```