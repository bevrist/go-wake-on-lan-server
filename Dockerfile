FROM golang:1 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-w -s" -o /wake-on-lan-server .


###############################
FROM scratch
COPY --from=builder /wake-on-lan-server /wake-on-lan-server
ENTRYPOINT ["/wake-on-lan-server"]

ENV SHARED_KEY=DefaultPassword
EXPOSE 80
