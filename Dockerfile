ARG CODENAME
ARG BUILD_TAG
ARG BUILD_DATE

FROM registry.drycc.cc/drycc/go-dev AS build
ARG LDFLAGS
ADD . /workspace
RUN export GO111MODULE=on \
  && cd /workspace \
  && pwd && ls \
  && init-stack go mod vendor \
  && init-stack go build \
    -ldflags "-X main.Revision=${BUILD_TAG} -X main.BuildDate=${BUILD_DATE}" \
    -buildmode=c-shared -o /var/lib/fluent-bit/out_drycc.so drycc.go


FROM registry.drycc.cc/drycc/base:${CODENAME}

ENV FLUENT_BIT_VERSION=2.1.8
ENV FLUENT_BIT_PLUGINS_PATH=/opt/drycc/fluent-bit/plugins

USER root
RUN install-stack fluent-bit ${FLUENT_BIT_VERSION}
RUN mkdir -p ${FLUENT_BIT_PLUGINS_PATH}
COPY --from=build /var/lib/fluent-bit/out_drycc.so ${FLUENT_BIT_PLUGINS_PATH}

ADD rootfs /

CMD ["/usr/local/bin/boot"]
