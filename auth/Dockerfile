FROM ubuntu:22.04

WORKDIR /app
COPY ./configs/cert.* ./
COPY ./configs/defaults.yaml ./
COPY ./cmd/main ./
RUN chmod +x /app/main
EXPOSE 8084
CMD ["/app/main"]