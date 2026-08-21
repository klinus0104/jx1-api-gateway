FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/api-gateway .

FROM alpine:3.22
RUN apk add --no-cache wget \
    && adduser -D -u 65532 gmapi
COPY --from=build /out/api-gateway /api-gateway
USER gmapi
EXPOSE 8080
ENTRYPOINT ["/api-gateway", "web-service", "start"]