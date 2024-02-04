FROM golang:alpine AS build_container
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
RUN go get go.mongodb.org/mongo-driver/mongo
COPY . .
RUN go build -o accommodation

FROM alpine:3.19
WORKDIR /usr/bin
COPY --from=build_container /app/accommodation .
EXPOSE 8080
ENTRYPOINT ["./accommodation"]
