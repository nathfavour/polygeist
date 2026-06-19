FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git
WORKDIR /src
COPY go.mod go.sum go.work go.work.sum ./
COPY anyisland/ anyisland/
COPY auracrab/ auracrab/
COPY vibeauracle/ vibeauracle/
COPY cmd/ cmd/
COPY pkg/ pkg/
RUN go build -ldflags "-s -w" -o /polygeist ./cmd/polygeist

FROM alpine:3.20
RUN apk add --no-cache ca-certificates git docker-cli
COPY --from=builder /polygeist /usr/local/bin/polygeist
RUN mkdir -p /run/agentic /workspace
ENV AGENTIC_RUN_DIR=/run/agentic
WORKDIR /workspace
ENTRYPOINT ["polygeist"]
CMD ["--version"]
