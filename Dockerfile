FROM scratch
COPY serve /bin/serve
ENTRYPOINT ["/bin/serve"]
