FROM golang:alpine as build_container
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
RUN go get go.mongodb.org/mongo-driver/mongo
COPY . .
RUN go build -o notification

FROM alpine
WORKDIR /usr/bin
COPY --from=build_container /app/notification .
EXPOSE 8084
ENTRYPOINT ["./notification"]
