FROM golang:1.24-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /schema-mcp ./server

FROM alpine:3.20
COPY --from=build /schema-mcp /usr/local/bin/schema-mcp
ENTRYPOINT ["schema-mcp"]
