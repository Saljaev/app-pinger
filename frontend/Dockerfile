FROM node:22-alpine3.21 as build

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json ./
RUN npm install --frozen-lockfile

COPY frontend .
COPY .env .

RUN npm run build

FROM nginx:1.27-alpine3.21

WORKDIR /etc/nginx

COPY --from=build /app/build /usr/share/nginx/html

COPY --from=build /app/nginx.conf.template /etc/nginx/nginx.conf.template

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
