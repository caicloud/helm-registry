#
# Copyright 2017 The Caicloud Authors.
#

FROM cargo.caicloudprivatetest.com/caicloud/debian:jessie

LABEL maintainer="Wei Guo <guowei@caicloud.io>"

RUN mkdir /data

COPY build/registry/config.yaml /config.yaml
COPY bin/registry /registry

ENTRYPOINT ["/registry"]
CMD ["serve","-c","/config.yaml"]

