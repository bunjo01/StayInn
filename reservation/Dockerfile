# Build stage
FROM golang:alpine AS build_container
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o reservation

# Final stage
FROM alpine:3.19
WORKDIR /app
COPY --from=build_container /app/reservation /usr/bin
EXPOSE 8082
ENTRYPOINT ["reservation"]

