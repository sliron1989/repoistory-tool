package templates

func ServiceDockerfile() string {
	return `FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /service .

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
COPY --from=builder /service /service
EXPOSE 8080
ENTRYPOINT ["/service"]
`
}
