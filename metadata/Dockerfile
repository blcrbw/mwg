FROM ubuntu:22.04

WORKDIR /app
COPY ./configs/cert.* ./
COPY ./configs/defaults.yaml ./
COPY ./cmd/main ./
RUN chmod +x ./main
EXPOSE 8081
RUN chown 1000:1000 ./main
USER 1000
CMD ["/app/main"]