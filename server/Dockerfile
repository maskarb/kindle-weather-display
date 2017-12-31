FROM golang:1.9.2 as builder

ADD main.go /go/

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -a -installsuffix cgo -o weather-script .

FROM alpine

WORKDIR /opt

RUN apk add --no-cache ca-certificates librsvg pngcrush ttf-dejavu

COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip
COPY --from=builder /go/weather-script /opt/
ADD weather-script.sh preprocess.svg /opt/

CMD ["/opt/weather-script.sh"]
