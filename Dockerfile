FROM golang:1.12


COPY . /go/src/camunda-engine-wrapper
WORKDIR /go/src/camunda-engine-wrapper

ENV GO111MODULE=on

RUN go build

EXPOSE 8080

CMD ./camunda-engine-wrapper