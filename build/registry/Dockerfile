FROM cargo.caicloud.xyz/library/debian:stretch

RUN mkdir /data

COPY build/registry/config.yaml /config.yaml
COPY bin/registry /registry

ENTRYPOINT ["/registry"]
CMD ["serve", "-c", "/config.yaml"]
