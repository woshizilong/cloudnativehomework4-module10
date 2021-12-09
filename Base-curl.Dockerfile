ARG ALPINE_TAG
FROM alpine:${ALPINE_TAG}
RUN apk add --update curl \
                     tini \
 && rm -rf /var/cache/apk/*