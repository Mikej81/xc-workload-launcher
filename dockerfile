FROM golang:1.19 AS Builder

ENV GO111MODULE=on \
    CGO_ENABLED=0

WORKDIR /app

# Copy the go mod and sum files first to leverage Docker cache layering
COPY go.mod ./
COPY *.html ./

RUN go mod download

COPY *.go ./
COPY *.sum ./

RUN go build -o workload-launcher -v .

FROM alpine:latest  

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/workload-launcher .
COPY --from=builder /app/*.html .

RUN chmod -R 755 .

EXPOSE 8080

CMD ["./workload-launcher"]