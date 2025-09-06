FROM ubuntu:22.04

WORKDIR /app
COPY ./configs/cert.* ./
COPY ./configs/defaults.yaml ./
COPY ./cmd/main ./
RUN chmod +x /app/main
RUN chown 1000:1000 ./main
USER 1000
EXPOSE 8084
CMD ["/app/main"]