FROM demisto/opencv:1.0.0.14169403

WORKDIR /

COPY executor.py /
RUN mkdir -p /app

CMD ["python3", "/executor.py"]