FROM public.ecr.aws/docker/library/golang:latest as builder
RUN mkdir -p /workspace
WORKDIR /workspace

COPY go.mod .
COPY go.sum .
RUN go mod download && go mod verify

COPY . .
# Use the ARG for GOARCH here
RUN CGO_ENABLED=0 GOOS=linux go build -o /app .

FROM public.ecr.aws/docker/library/alpine:3.19.0
COPY --from=builder /app /app
ENTRYPOINT ["/app"]
