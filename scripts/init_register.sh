docker pull grussorusso/serverledge-python310
docker tag grussorusso/serverledge-python310 localhost:5000/python310
docker push localhost:5000/python310
docker image rm grussorusso/serverledge-python310
docker image rm localhost:5000/python310