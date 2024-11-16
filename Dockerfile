FROM public.ecr.aws/docker/library/golang:latest as builder
RUN mkdir -p /workspace
WORKDIR /workspace

COPY go.mod .
COPY go.sum .
RUN go mod download && go mod verify

COPY . .
# Use the ARG for GOARCH here

# since I am using github.com/mattn/go-sqlite3 at the moment and its c-go I have to enable cgo
#RUN CGO_ENABLED=0 GOOS=linux go build -o /app .
RUN GOOS=linux go build -o app .

# having to use debian instead of alpine for the c libs that sqlite stuff requires
#FROM public.ecr.aws/docker/library/debian:bullseye-slim
FROM public.ecr.aws/docker/library/alpine:latest
RUN apk add --no-cache libc6-compat
COPY --from=builder /workspace/app /app
ENTRYPOINT ["/app"]
