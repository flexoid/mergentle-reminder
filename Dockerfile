FROM golang:1.20 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o mergentle-reminder

FROM gcr.io/distroless/base-debian11

COPY --from=builder /app/mergentle-reminder /mergentle-reminder

CMD ["/mergentle-reminder"]
