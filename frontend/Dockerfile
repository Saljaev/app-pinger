FROM node:22-alpine3.21 as builder

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json ./
RUN npm install --frozen-lockfile

COPY frontend .
COPY .env .

RUN npm run build

FROM nginx:1.27-alpine3.21

WORKDIR /etc/nginx

COPY --from=builder /app/build /usr/share/nginx/html

COPY --from=builder /app/nginx.conf.template /etc/nginx/nginx.conf.template

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
