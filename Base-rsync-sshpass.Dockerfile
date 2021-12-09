ARG ALPINE_TAG
FROM alpine:${ALPINE_TAG}
RUN apk update \ 
 && apk upgrade \ 
 && apk add --no-cache \ 
            curl \
            tini \
            git \
            rsync \ 
            openssh-client \ 
            openssh \ 
            sshpass \ 
            ca-certificates \ 
 && update-ca-certificates \ 
 && rm -rf /var/cache/apk/* 