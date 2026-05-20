FROM golang:1.26.2-alpine AS builder
WORKDIR /app
COPY . .

RUN apk add --no-cache make git  && \
    rm -rf build/ && mkdir -p build/ && \
    go mod tidy && \
    make build


FROM scratch AS final
WORKDIR /app

LABEL Author="duakc"
LABEL Description="A Lightweight DDNS(Dynamic DNS) Prog."
LABEL Name="lightddns"

COPY --from=builder /app/build/lightddns lightddns
COPY --from=builder /etc/ssl /etc/ssl

ENV PATH="/app"

CMD ["/app/lightddns","run","-c","/etc/lightddns.yaml","-D","/data"]