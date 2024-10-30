FROM python:3.12-alpine

LABEL org.opencontainers.image.authors="Michael Krug <michi.krug@gmail.com>"
LABEL org.opencontainers.image.description="An implementation of a Prometheus exporter for SmartThings Devices"
LABEL org.opencontainers.image.source=https://github.com/michikrug/smartthings_exporter
LABEL org.opencontainers.image.licenses=GPL-3.0

RUN apk update && apk add py3-pip

ADD requirements.txt /requirements.txt
RUN pip install -r /requirements.txt

ADD smartthings_exporter.py /smartthings_exporter.py

EXPOSE 9090

CMD [ "python", "/smartthings_exporter.py" ]
