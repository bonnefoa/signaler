#!/bin/bash

openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -nodes -subj "/C=US/ST=Oregon/L=Portland/O=Company Name/OU=Org/CN=www.example.com"
