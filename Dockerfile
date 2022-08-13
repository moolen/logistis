FROM golang:1.17 AS build

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

WORKDIR /work
COPY . /work

RUN --mount=type=cache,target=/root/.cache/go-build,sharing=private \
  go build -o bin/logistis ./cmd/webhook

FROM alpine AS run

RUN apk --no-cache add curl
COPY --from=build /work/bin/logistis /usr/local/bin/

CMD ["logistis"]
