FROM golang:1.23.2-alpine3.20 as builder
WORKDIR /tsp
RUN apk update && apk upgrade --available && sync && apk add --no-cache --virtual .build-deps
COPY . .
RUN go build -ldflags="-w -s" .
FROM alpine:3.20.3
RUN apk update && apk upgrade --available && sync
COPY --from=builder /tsp/tsp /tsp
COPY --from=builder /tsp/incidents.html /incidents.html
COPY --from=builder /tsp/checks.yaml /checks.yaml
ENTRYPOINT ["/tsp"]