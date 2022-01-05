FROM python:3.10.1-alpine3.14
RUN apk add --no-cache git
WORKDIR /app
COPY requirements.txt backup.py entrypoint.sh config.json ./
RUN pip install --no-cache-dir -r requirements.txt
ENTRYPOINT [ "./entrypoint.sh" ]
