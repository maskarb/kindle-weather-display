FROM golang:1.15 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY main.go main.go
COPY preprocess.svg preprocess.svg

RUN CGO_ENABLED=0 GOOS=linux go build -o weather-script .

FROM alpine

WORKDIR /opt

RUN apk add --no-cache ca-certificates librsvg && mkdir out

COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip
COPY --from=builder /workspace/weather-script /opt/
COPY --from=builder /workspace/preprocess.svg /opt/


CMD ["/opt/weather-script"]
