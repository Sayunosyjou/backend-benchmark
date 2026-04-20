FROM openresty/openresty:1.25.3.2-0-bookworm
RUN /usr/local/openresty/bin/opm get SkyLothar/lua-resty-jwt && /usr/local/openresty/bin/opm get openresty/lua-resty-redis
COPY nginx.conf /usr/local/openresty/nginx/conf/nginx.conf
COPY lua /etc/nginx/lua
EXPOSE 8088
CMD ["/usr/local/openresty/bin/openresty", "-g", "daemon off;"]
