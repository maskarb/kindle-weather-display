FROM golang:1.15 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY main.go main.go
COPY preprocess.svg preprocess.svg

RUN CGO_ENABLED=0 GOOS=linux go build -o kindle-server .

FROM alpine

WORKDIR /opt

RUN apk add --no-cache tzdata ttf-dejavu librsvg pngcrush

COPY --from=builder /workspace/kindle-server /opt/
COPY --from=builder /workspace/preprocess.svg /opt/

ENTRYPOINT ["/opt/kindle-server"]
