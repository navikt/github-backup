FROM python:3.10.1-alpine3.14
RUN apk add --no-cache \
    git \
    openssh \
    jq \
    bash
RUN pip install --upgrade pip

RUN echo "Host *" > /etc/ssh/ssh_config
RUN echo "  IdentityFile /home/backup/.ssh/id" >> /etc/ssh/ssh_config
RUN echo "  StrictHostKeyChecking no" >> /etc/ssh/ssh_config

RUN adduser -D backup -u 1069
WORKDIR /home/backup
USER backup

ENV PATH="/home/backup/.local/bin:${PATH}"
COPY --chown=backup:backup requirements.txt requirements.txt
RUN pip install --user -r requirements.txt

COPY --chown=backup:backup backup.py config.json entrypoint.sh ./

ENTRYPOINT [ "/home/backup/entrypoint.sh" ]
