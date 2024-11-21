FROM golang:1.23-alpine AS build
WORKDIR /build
COPY . .
RUN go mod download
RUN go build -o sim cmd/server/main.go

FROM busybox
WORKDIR /app
COPY --from=build /build/sim .

EXPOSE 7547
ENTRYPOINT ["/app/sim"]
