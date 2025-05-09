# 🗂️ Local S3 Server with Go + MinIO + PM2 + Token Auth + NGINX

## 📦 Описание

Этот проект поднимает локальный S3-совместимый сервер с помощью **MinIO**, обрабатывает загрузку и скачивание файлов через **Go**, управляется через **PM2**, защищает загрузку **токеном из .env**, и проксируется через **NGINX** с доменом `https://yourdomain`.

---

## 📁 Структура проекта

```
project/
├── go-server/             # Go-код
│   ├── main.go
│   ├── go.mod
│   └── .env               # Переменные окружения
├── minio-data/            # Данные MinIO
├── start-minio.sh         # Обёртка для MinIO с переменными
├── pm2.config.js          # Конфигурация PM2
└── README.md              # Этот файл
```

---

## ⚙️ Установка

### 1. Установи зависимости

```bash
sudo apt update
sudo apt install golang nodejs npm nginx curl -y
npm install -g pm2
```

---

## 🚀 Запуск

### 1. Создай `.env`

```env
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
BASE_URL=https://localhost:3333/files
UPLOAD_TOKEN=supersecrettoken
```

---

### 2. Собери Go-сервер

```bash
cd go-server
go mod tidy
go build -o ../go-server
```

---

### 3. Запусти через PM2

```bash
pm2 start pm2.config.js
pm2 save
```

Или вручную:

```bash
pm2 start ./go-server --name go-server --interpreter none -- --port 3333
pm2 start ./start-minio.sh --name minio
```

---

### 4. Настрой NGINX

Пример конфигурации:

```nginx
server {
    listen 80;
    server_name yourdomain;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name yourdomain;

    ssl_certificate     /etc/ssl/domain/self.crt;
    ssl_certificate_key /etc/ssl/domain/self.key;

    location / {
        proxy_pass http://localhost:3333;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## 🔐 Авторизация

### Загрузка защищена токеном

```bash
curl -F "file=@test.jpg" \
  -H "Authorization: Bearer supersecrettoken" \
  https://yourdomain/upload
```

---

## 🧪 Доступ к файлам

```bash
curl https://yourdomain/files/<hash>
```

---

## 🛠 PM2 команды

```bash
pm2 list
pm2 logs go-server
pm2 restart all
pm2 save
```

---

## 📄 Лицензия

MIT
