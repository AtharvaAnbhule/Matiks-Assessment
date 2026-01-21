 
FROM golang:1.21-alpine AS builder

 
RUN apk add --no-cache git make gcc musl-dev

 
WORKDIR /app

 
COPY go.mod go.sum ./

 
RUN go mod download

 
COPY . .

 
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o app main.go

 
FROM alpine:latest

 
RUN apk --no-cache add ca-certificates curl

 
WORKDIR /root/

 
COPY --from=builder /app/app .

 
EXPOSE 8080

 
HEALTHCHECK --interval=10s --timeout=5s --retries=5 \
    CMD curl -f http://localhost:8080/health || exit 1

 
CMD ["./app"]
