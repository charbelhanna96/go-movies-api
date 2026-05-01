FROM   postgres:18-alpine
RUN    apk update && apk upgrade --no-cache libexpat
EXPOSE 5432
ENV    POSTGRES_PASSWORD=password123 \
       POSTGRES_DB=movie \
       POSTGRES_USER=postgres
COPY   migrate.sql /docker-entrypoint-initdb.d/
