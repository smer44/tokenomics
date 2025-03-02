FROM mirror.gcr.io/library/golang:alpine as build

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOCACHE=/root/.cache/go-build
	
WORKDIR /build
COPY . .

RUN go test ./... && go build -o app ./cmd/tokenomics-server

# FROM gcr.io/distroless/static:nonroot
FROM alpine:latest
WORKDIR /
COPY --from=build /build/app .
EXPOSE 8080

ENTRYPOINT ["/app", "--port", "8080", "--host", "0.0.0.0"] 
