FROM scratch

COPY ./minecraft-skinsystem /app/

ENTRYPOINT ["/app/minecraft-skinsystem"]
CMD ["serve"]
