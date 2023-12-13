FROM golang:alpine as build_container
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o notification

FROM alpine
COPY --from=build_container /app/notification /usr/bin
EXPOSE 8084
ENTRYPOINT ["notification"]
