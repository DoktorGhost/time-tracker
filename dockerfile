# Используем официальный образ PostgreSQL
FROM postgres:latest

ENV POSTGRES_USER admin
ENV POSTGRES_PASSWORD admin
ENV POSTGRES_DB admin

# Запуск PostgreSQL и вашего приложения при запуске контейнера
CMD ["postgres"]