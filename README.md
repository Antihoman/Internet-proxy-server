# Internet-proxy-server

# Запуск прокси-сервера и web api на хосте, бд в докере

- переименовать docker-compose-mongo.yml в docker-compose.yml
- docker compose up
- make -B build
- ./build/proxy/out
- ./build/webapi/out

# Запуск прокси-сервера, web api и бд в докере

- docker compose up

# Cертификатs

- sudo cp .mitm/ca-cert.pem /usr/local/share/ca-certificates/ca-cert.crt
- sudo update-ca-certificates

# Поиск уязвимостей
- curl -x 127.0.0.1:8080 -v http://212.233.91.39/?name=kirill
- request.get_params.name = kirill и response.text_body = Hello, kirill!