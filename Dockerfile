FROM golang:1.20.7 AS builder
COPY . /root/project
WORKDIR /root/project
RUN  go build -mod vendor -o pagerdutybot cmd/pagerdutybot/main.go 

FROM gcr.io/distroless/base
LABEL name=jeffgtxjava@gmail.com
WORKDIR /app
COPY --from=builder /root/project/pagerdutybot /usr/local/bin/pagerdutybot
ENTRYPOINT ["/usr/local/bin/pagerdutybot"]
