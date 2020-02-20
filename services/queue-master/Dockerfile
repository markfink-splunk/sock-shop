FROM java:openjdk-8-alpine

ENV	SERVICE_USER=myuser \
	SERVICE_UID=10001 \
	SERVICE_GROUP=mygroup \
	SERVICE_GID=10001

RUN	addgroup -g ${SERVICE_GID} ${SERVICE_GROUP} && \
	adduser -g "${SERVICE_NAME} user" -D -H -G ${SERVICE_GROUP} -s /sbin/nologin -u ${SERVICE_UID} ${SERVICE_USER} && \
	apk add --update libcap && \
	mkdir /lib64 && \
	ln -s /usr/lib/jvm/java-1.8-openjdk/jre/lib/amd64/server/libjvm.so /lib/libjvm.so && \
	ln -s /usr/lib/jvm/java-1.8-openjdk/lib/amd64/jli/libjli.so /lib/libjli.so && \
	setcap 'cap_net_bind_service=+ep' $(readlink -f $(which java))

WORKDIR /usr/src/app
COPY ./docker/queue-master/*.jar ./app.jar

# SIGNALFX UAPM
COPY ./signalfx-tracing.jar ./signalfx-tracing.jar

RUN	chown -R ${SERVICE_USER}:${SERVICE_GROUP} ./*.jar

USER ${SERVICE_USER}

# Add "-javaagent:signalfx-tracing.jar" to the JAVA_OPTS env variable.
# JAVA_OPTS needs to be part of the command line when run using the "shell"
# form as below. I can't remember now why I needed to use shell form here, but
# I did.
ENTRYPOINT java $JAVA_OPTS -jar ./app.jar --port=80
