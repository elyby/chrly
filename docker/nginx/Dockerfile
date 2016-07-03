FROM nginx:1.11

COPY nginx.conf /etc/nginx/nginx.conf
COPY skinsystem.ely.by.conf /etc/nginx/conf.d/default.conf

RUN mkdir -p /data/cache \
             /data/logs \
 && chown nginx:nginx -R /data
