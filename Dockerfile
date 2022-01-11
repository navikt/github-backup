FROM python:3.10.1-alpine3.14
RUN apk add --no-cache \
    git \
    openssh \
    jq
RUN pip install --upgrade pip

RUN adduser -D backup -u 1069 -h /tmp/backup
USER backup
WORKDIR /tmp/backup

RUN mkdir -p ~/.ssh && \
    chmod 0700 ~/.ssh

ENV PATH="/tmp/backup/.local/bin:${PATH}"
COPY --chown=backup:backup requirements.txt requirements.txt
RUN pip install --user -r requirements.txt

COPY --chown=backup:backup backup.py config.json entrypoint.sh ./

ENTRYPOINT [ "./entrypoint.sh" ]
