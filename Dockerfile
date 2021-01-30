FROM golang as builder

WORKDIR /go/src/app/

ENV GO111MODULE=on

RUN groupadd -g 10001 myapp \
    && useradd -u 10001 -g myapp myapp

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /go/bin/app ./main.go; \
    chmod -R 777 /go/src/app/

FROM scratch

COPY --from=builder /go/bin/app /go/bin/app
COPY --from=builder /etc/group /etc/group
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /go/src/app/private.key /go/bin/

WORKDIR /go/bin
USER myapp

CMD ["/go/bin/app"]
