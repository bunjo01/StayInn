# Build stage
FROM golang:alpine as build_container
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o reservation

# Final stage
FROM alpine
WORKDIR /app
COPY --from=build_container /app/reservation /usr/bin
EXPOSE 8080
ENTRYPOINT ["reservation"]

