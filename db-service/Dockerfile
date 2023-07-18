FROM golang:1.19.8
WORKDIR /
COPY src/* /
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o database-server .

# TODO test alpine
FROM alpine
EXPOSE 7272

WORKDIR /program
COPY entrypoint.sh .
COPY --from=0 database-server .

CMD ["sh", "entrypoint.sh"]