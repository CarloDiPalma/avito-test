FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY ./backend /app

RUN go mod tidy
RUN go build -o main .

FROM alpine:latest

WORKDIR /root/


COPY --from=builder /app/main .

ENV SERVER_ADDRESS=0.0.0.0:8080
ENV POSTGRES_CONN=postgres://cnrprod1725742191-team-77945:cnrprod1725742191-team-77945@rc1b-5xmqy6bq501kls4m.mdb.yandexcloud.net:6432/cnrprod1725742191-team-77945?target_session_attrs=read-write
ENV POSTGRES_USERNAME=cnrprod1725742191-team-77945
ENV POSTGRES_PASSWORD=cnrprod1725742191-team-77945
ENV POSTGRES_HOST=rc1b-5xmqy6bq501kls4m.mdb.yandexcloud.net
ENV POSTGRES_PORT=6432
ENV POSTGRES_DATABASE=cnrprod1725742191-team-77945

EXPOSE 8080

CMD ["./main"]