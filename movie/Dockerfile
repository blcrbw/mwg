FROM ubuntu:22.04

WORKDIR /app
COPY ./configs/cert.* ./
COPY configs/defaults.yaml ./
COPY ./cmd/main ./
RUN chmod +x ./main
EXPOSE 8083
CMD ["/app/main"]