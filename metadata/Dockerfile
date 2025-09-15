FROM ubuntu:22.04

WORKDIR /app
COPY ./configs/cert.* ./
COPY ./configs/defaults.yaml ./
COPY ./cmd/main ./
RUN chmod +x ./main
USER 1000
CMD ["/app/main"]