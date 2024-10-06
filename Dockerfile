ARG BUSYBOX_IMAGE
FROM ${BUSYBOX_IMAGE}

COPY h3d-drone-emulator /h3d-drone-emulator
RUN chmod +x /h3d-drone-emulator

ENTRYPOINT [ "/h3d-drone-emulator" ]
