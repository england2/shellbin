FROM golang:1.19.8
WORKDIR /
COPY src/* /
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o ncServer .

# TODO test alpine
FROM alpine
EXPOSE 6262

WORKDIR /program
COPY entrypoint.sh .
COPY --from=0 server .

CMD ["sh", "entrypoint.sh"]