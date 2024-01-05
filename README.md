## Build the backend with:

``` bash
docker buildx build -t docdiff --target prod .
```

## RUN the backend with:
``` bash
docker run --env-file=.env -d -p 5000:5000 --name=docdiff docdiff:latest
```

# PS: MAKE SURE THE .env is on the root directory of the project:
``` bash
  .
  ├── frontend..
  ├── backend
  │   ├── docfiff
  │   ├── Examples
  │   │   └── Response.json
  │   ├── files
  │   ├── go.mod
  │   ├── go.sum
  │   ├── Makefile
  │   └── src
  │       ├── api
  │       │   ├── api.go
  │       │   └── pdf.go
  │       ├── db
  │       │   └── database.go
  │       └── main.go
  ├── Dockerfile
  ├── .env
  ├── .dockerignore
  └── README.md
```
