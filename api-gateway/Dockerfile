FROM nginx:latest

COPY ./api_gateway.conf /etc/nginx/conf.d/api_gateway.conf
COPY ./localhost.crt /etc/ssl/certs/localhost.crt
COPY ./localhost.csr /etc/ssl/csr/localhost.csr
COPY ./localhost.ext /etc/ssl/certs/localhost.ext
COPY ./localhost.key /etc/ssl/private/localhost.key

RUN rm /etc/nginx/conf.d/default.conf
