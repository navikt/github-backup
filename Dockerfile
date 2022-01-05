FROM python:3.10.1-alpine3.14
RUN apk add --no-cache git

RUN adduser -D backup
USER backup
WORKDIR /home/backup

COPY --chown=backup:backup requirements.txt requirements.txt
RUN pip install --user -r requirements.txt

COPY --chown=backup:backup backup.py config.json entrypoint.sh ./

ENTRYPOINT [ "./entrypoint.sh" ]
