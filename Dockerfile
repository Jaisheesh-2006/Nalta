FROM golang:1.24-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /nalta .

FROM alpine:3.20
COPY --from=build /nalta /usr/local/bin/nalta
ENTRYPOINT ["nalta"]
