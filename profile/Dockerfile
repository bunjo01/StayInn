FROM golang:alpine as build_container
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o profile

FROM alpine
COPY --from=build_container /app/profile /usr/bin
EXPOSE 8083
ENTRYPOINT ["profile"]
