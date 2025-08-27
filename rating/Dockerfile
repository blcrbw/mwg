FROM ubuntu:22.04

WORKDIR /app
COPY ./cmd/main ./
COPY ./configs/defaults.yaml ./
RUN chmod +x /app/main
EXPOSE 8082
CMD ["/app/main"]