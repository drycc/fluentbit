ARG CODENAME
ARG BUILD_TAG
ARG BUILD_DATE

FROM registry.drycc.cc/drycc/go-dev AS build
ADD . /workspace
RUN export GO111MODULE=on \
  && cd /workspace \
  && pwd && ls \
  && init-stack go mod vendor \
  && init-stack go build \
    -ldflags "-X main.Revision=${BUILD_TAG} -X main.BuildDate=${BUILD_DATE}" \
    -buildmode=c-shared -o /var/lib/fluent-bit/out_drycc.so plugin/out_drycc.go


FROM registry.drycc.cc/drycc/base:${CODENAME}

ENV DRYCC_UID=1001 \
  DRYCC_GID=1001 \
  FLUENT_BIT_VERSION=3.2.3 \
  DRYCC_HOME_DIR=/opt/drycc \
  FLUENT_BIT_PLUGINS_PATH=${DRYCC_HOME_DIR}/fluent-bit/plugins
  
RUN groupadd drycc --gid ${DRYCC_GID} \
  && useradd drycc -u ${DRYCC_UID} -g ${DRYCC_GID} -s /bin/bash -m -d ${DRYCC_HOME_DIR}

RUN install-stack fluent-bit ${FLUENT_BIT_VERSION} \
  && mkdir -p ${FLUENT_BIT_PLUGINS_PATH} \
  && chown -R ${DRYCC_UID}:${DRYCC_GID} ${DRYCC_HOME_DIR}/fluent-bit

COPY --chown=${DRYCC_UID}:${DRYCC_GID} --from=build /var/lib/fluent-bit/out_drycc.so ${FLUENT_BIT_PLUGINS_PATH}

ADD rootfs /

USER ${DRYCC_UID}

CMD ["/usr/local/bin/boot"]
