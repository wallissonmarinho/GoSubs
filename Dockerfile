FROM golang:1.25-alpine AS build
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /gosubs ./cmd/gosubs

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /gosubs /usr/local/bin/gosubs
ENV PORT=8090
EXPOSE 8090
ENTRYPOINT ["/usr/local/bin/gosubs"]
