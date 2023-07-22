FROM golang:1.19.8
WORKDIR /
COPY src/* /
RUN go mod tidy
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o db-service .

# TODO test alpine
FROM alpine
EXPOSE 7272

WORKDIR /program
COPY --from=0 db-service .
COPY entrypoint.sh .

CMD ["sh", "entrypoint.sh"]