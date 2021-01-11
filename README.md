# CQueue

Cqueue is an uncomplicated queue service for batch processing powered by Docker.

## Create Containers

```bash
docker-compose up -d
```

## Examples

### Default execution:
```bash
curl -H 'Content-Type: application/json' -X POST -d'{"image":"ubuntu", "cmd":["echo", "test msg"]}' http://localhost:8080/task
```

### Local execution
```bash
curl -H 'Content-Type: application/json' -X POST -d'{"type":"local", "cmd":["echo", "test msg"]}' http://localhost:8080/task
```

### Batch execution
```bash
curl -H 'Content-Type: application/json' -X POST -d'{"type":"batch", "start":"1" , "stop":"10", "cmd":["echo", "run {{.}}.cfg"]}' http://localhost:8080/task
```

### Check status (The following job states are available PENDING, RECEIVED, STARTED, RETRY, SUCCESS, FAILURE.)

```bash
curl -X GET http://localhost:8080/task/$(myUniqId)
```

### Fetch result (Stdout of the job.)

```bash
curl -X GET http://localhost:8080/task/$(myUniqId)/result
```

### Purge task

```bash
curl -X DELETE http://localhost:8080/task/$(myUniqId)
```

# Binary arguments

## Frontend
- "config" - Path of machinery config file. (https://github.com/RichardKnop/machinery#configuration)
- "http" - Address to listen for HTTP requests on. (default: "0.0.0.0:8080")

## Worker
- "config" - Path of machinery config file. (https://github.com/RichardKnop/machinery#configuration)
- "concurrency" - Number of concurrent jobs/worker. (default: single)
- "tag" - Tag of the worker instance. (default: random integer)
- "timeout" - Exit after timeout. (disabled by default)
