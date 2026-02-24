FROM golang:1.24.2-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /out/delayednotifier ./cmd


FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /out/delayednotifier /app/delayednotifier
COPY internal/web/static /app/internal/web/static

EXPOSE 8080

CMD ["/app/delayednotifier"]
