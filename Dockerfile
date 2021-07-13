FROM golang:1.16 AS builder

COPY . /go/src/app
WORKDIR /go/src/app

ENV GO111MODULE=on

RUN CGO_ENABLED=0 GOOS=linux go build -o app
RUN CGO_ENABLED=0 GOOS=linux go build -o cleanup ./cmd/cleanup
RUN CGO_ENABLED=0 GOOS=linux go build -o addshard ./cmd/addshard
RUN CGO_ENABLED=0 GOOS=linux go build -o removeshard ./cmd/removeshard

RUN git log -1 --oneline > version.txt

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /go/src/app/app .
COPY --from=builder /go/src/app/cleanup .
COPY --from=builder /go/src/app/addshard .
COPY --from=builder /go/src/app/removeshard .
COPY --from=builder /go/src/app/config.json .
COPY --from=builder /go/src/app/version.txt .

# add command completion for cleanup
RUN apk add bash
RUN apk add bash-completion
COPY --from=builder /go/src/app/cmd/cleanup/cleanup-completion.bash /usr/share/bash-completion/completions/cleanup
COPY --from=builder /go/src/app/cmd/cleanup/cleanup-completion.bash .profile
COPY --from=builder /go/src/app/cmd/cleanup/cleanup-completion.bash .bashrc


EXPOSE 8080

ENTRYPOINT ["./app"]