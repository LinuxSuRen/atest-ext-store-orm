ARG GO_BUILDER=docker.io/library/golang:1.22
ARG BASE_IMAGE=docker.io/library/alpine:3.12

FROM ${GO_BUILDER} AS builder

ARG VERSION
ARG GOPROXY
WORKDIR /workspace
COPY . .

RUN GOPROXY=${GOPROXY} go mod download
RUN GOPROXY=${GOPROXY} CGO_ENABLED=0 go build -ldflags "-w -s" -o atest-store-orm .

FROM ${BASE_IMAGE}

LABEL org.opencontainers.image.source=https://github.com/LinuxSuRen/atest-ext-store-orm
LABEL org.opencontainers.image.description="ORM database Store Extension of the API Testing."

COPY --from=builder /workspace/atest-store-orm /usr/local/bin/atest-store-orm

CMD [ "atest-store-orm" ]
