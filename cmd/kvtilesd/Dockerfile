FROM gcr.io/distroless/base

WORKDIR /root/
COPY kvtilesd .
COPY grpc_health_probe .
ADD static /root/static
EXPOSE 6666 8088 9200
ENTRYPOINT ["./kvtilesd"]
