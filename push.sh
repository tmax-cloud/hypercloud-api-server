#!/bin/sh
img=192.168.9.12:5000/hypercloud-api-server:v0.24
docker rmi $img 
docker build -t $img  . 
docker push $img 
