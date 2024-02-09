FROM golang:1.22.0-alpine3.19 AS build
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o app .
FROM alpine:3.19.1
WORKDIR /root/
COPY --from=build /app/app .
EXPOSE 3000

CMD ["./app"]
