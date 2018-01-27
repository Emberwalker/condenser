# Based on golang sample and Redis https://github.com/docker-library/redis/blob/master/4.0/Dockerfile

FROM golang:1

# Build
LABEL maintainer="Arkan <arkan@drakon.io>"
LABEL source="https://github.com/emberwalker/condenser"

RUN groupadd -r condenser && useradd --no-log-init -r -g condenser condenser

WORKDIR /go/src/github.com/emberwalker/condenser
COPY . .
RUN go install -v ./... && mkdir /condenser && chown condenser:condenser /condenser

# Run
WORKDIR /condenser
USER condenser:condenser
VOLUME ["/condenser"]
EXPOSE 8000
CMD ["/go/bin/condenser"]