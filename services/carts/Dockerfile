FROM java:openjdk-8-alpine

WORKDIR /usr/src/app
COPY ./target/*.jar ./app.jar

# SIGNALFX UAPM AUTO-INSTRUMENTATION
COPY signalfx-tracing.jar .

# Add "-javaagent:signalfx-tracing.jar" to the JAVA_OPTS env variable.
# JAVA_OPTS needs to be part of the command line when run using the "shell"
# form as below. I can't remember now why I needed to use shell form here, but
# I did.
ENTRYPOINT java $JAVA_OPTS -jar ./app.jar --port=8080
