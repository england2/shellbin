FROM golang:1.19.8
WORKDIR /build
COPY src/ ./
RUN go mod download
RUN CGO_ENABLED=0 go build -o webserver .

FROM alpine
EXPOSE 4747

WORKDIR /program
COPY entrypoint.sh .
COPY --from=0 /build/webserver .
COPY --from=0 /build/assets/templates ./assets/templates

CMD ["sh", "entrypoint.sh"]
