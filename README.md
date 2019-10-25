# Intro
Cqueue is an uncomplicated queue service for batch processing powered by docker.

# Build and compile
## Requirements

Go version => '1.8'. Check it, plase!
```bash
$ go version
```

Correct GOPATH. Check it, plase!
```bash
$ echo $GOPATH
```

Setup glide
- https://github.com/Masterminds/glide

## Download source code
with go get
```bash
$ go get gitlab.com/lpds-public/cqueue
```
or git clone
```bash
$ git clone git@gitlab.com:lpds-public/cqueue.git $GOPATH/src/gitlab.com/lpds-public/cqueue
```
## Install dependencies
```bash
$ cd $GOPATH/src/gitlab.com/lpds-public/cqueue
$ glide up
```

## Compile
```bash
mkdir -p bin && cd bin && find ../cmd -mindepth 1 -maxdepth 1 | xargs -n1 go build -v && cd ..
```

## Create Containers
```bash
# Launch frontend, worker, rabbitmq and redis servers in docker containers.
# Don't forget to rebuild the binaries and containers after you change something in the source code.
mkdir -p bin && cd bin && find ../cmd -mindepth 1 -maxdepth 1 | GOOS=linux xargs -n1 go build -v && cd ..
docker-compose build
docker-compose up -d
```

# Testing. Feed the queue:
Push task to docker executor
```bash
curl -H 'Content-Type: application/json' -X POST -d'{"image":"ubuntu", "cmd":["echo", "hello Docker"]}' http://localhost:8080/task
```

Push task to local executor
```bash
curl -H 'Content-Type: application/json' -X POST -d'{"type":"local", "cmd":["echo", "hello world"]}' http://localhost:8080/task
```

Check status (The following job states are available PENDING, RECEIVED, STARTED, RETRY, SUCCESS, FAILURE.)
```bash
curl -X GET http://localhost:8080/task/$(myUniqId)
```

Fetch result (Stdout of the job.)
```bash
curl -X GET http://localhost:8080/task/$(myUniqId)/result
```

Purge task
```bash
curl -X DELETE http://localhost:8080/task/$(myUniqId)
```

# Binary arguments
## Frontend
- "http" - Address to listen for HTTP requests on. (default: "0.0.0.0:8080")
- "config" - Path of machinery config file. (https://github.com/RichardKnop/machinery#configuration)

## Worker
- "config" - Path of machinery config file. (https://github.com/RichardKnop/machinery#configuration)
- "concurrency" - Number of concurrent jobs/worker. (default: single)
- "tag" - Tag of the worker instance. (default: random integer)
- "timeout" - Exit after timeout. (disabled by default)
