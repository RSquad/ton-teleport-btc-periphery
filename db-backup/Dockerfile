FROM alpine:3.14

RUN apk add --no-cache postgresql-client aws-cli bash gzip

COPY backup.sh /scripts/backup.sh
RUN chmod +x /scripts/backup.sh

WORKDIR /scripts

ENTRYPOINT ["/bin/bash", "/scripts/backup.sh"]
