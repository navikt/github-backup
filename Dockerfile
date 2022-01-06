FROM python:3.10.1-alpine3.14
RUN apk add --no-cache \
    git \
    openssh
RUN pip install --upgrade pip

RUN adduser -D backup
USER backup
WORKDIR /home/backup

RUN mkdir -p ~/.ssh && \
    chmod 0700 ~/.ssh

ENV PATH="/home/backup/.local/bin:${PATH}"
COPY --chown=backup:backup requirements.txt requirements.txt
RUN pip install --user -r requirements.txt

COPY --chown=backup:backup backup.py config.json entrypoint.sh ./

ENTRYPOINT [ "./entrypoint.sh" ]
