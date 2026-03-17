FROM golang:1.19.8
WORKDIR /
COPY src/* /
RUN go mod download
RUN CGO_ENABLED=0 go build -o ncserver .

# TODO test alpine
FROM alpine
EXPOSE 6262

WORKDIR /program
COPY entrypoint.sh .
COPY --from=0 ncserver .

CMD ["sh", "entrypoint.sh"]
