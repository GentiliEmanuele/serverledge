FROM tensorflow/tensorflow:2.16.1-jupyter

WORKDIR /

COPY executor.py /
RUN mkdir -p /app

CMD ["python3", "/executor.py"]