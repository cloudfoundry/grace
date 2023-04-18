FROM busybox:latest

ENV CUSTOM_GRACE_ENV my-docker-configured-env
CMD ["-chatty"]
ENTRYPOINT ["/grace"]

WORKDIR /tmp

EXPOSE 8080 9999

COPY grace /grace
RUN chmod a+x /grace

LABEL org.cloudfoundry.grace.dockerfile.url="https://github.com/cloudfoundry/grace/blob/main/Dockerfile"
LABEL org.cloudfoundry.grace.task.url="https://github.com/cloudfoundry/grace/blob/main/ci/build-grace-binary.yml"
LABEL org.cloudfoundry.grace.notes.md="Used by diego-release within \
code.cloudfoundry.org/inigo \
code.cloudfoundry.org/vizzini \
"
