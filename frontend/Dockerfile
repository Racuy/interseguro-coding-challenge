FROM nginx:1.27-alpine
COPY index.html style.css app.js /usr/share/nginx/html/
COPY docker-entrypoint.d/40-generate-config.sh /docker-entrypoint.d/40-generate-config.sh
RUN chmod +x /docker-entrypoint.d/40-generate-config.sh
EXPOSE 80
