FROM golang:1.20.5 as builder
RUN mkdir -p /workspace
WORKDIR /workspace

COPY go.mod .
COPY go.sum .
RUN go mod download && go mod verify

COPY . .
WORKDIR /workspace/test_tornado
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app .

FROM scratch
COPY --from=builder /app /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/app"]
