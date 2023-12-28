FROM golang:1.20 AS builder

LABEL org.opencontainers.image.source=https://github.com/LinuxSuRen/atest-ext-store-git
LABEL org.opencontainers.image.description="Git Store Extension of the API Testing."

ARG VERSION
ARG GOPROXY
WORKDIR /workspace
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY go.mod go.mod
COPY go.sum go.sum
COPY main.go main.go
COPY README.md README.md

RUN GOPROXY=${GOPROXY} go mod download
RUN GOPROXY=${GOPROXY} CGO_ENABLED=0 go build -ldflags "-w -s" -o atest-store-git .

FROM alpine:3.12

COPY --from=builder /workspace/atest-store-git /usr/local/bin/atest-store-git

CMD [ "atest-store-git" ]
