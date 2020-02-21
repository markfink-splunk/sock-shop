FROM weaveworksdemos/msd-java:jre-latest
#FROM java:openjdk-8-alpine

WORKDIR /usr/src/app
COPY orders.jar ./app.jar

# SIGNALFX UAPM
COPY signalfx-tracing.jar ./signalfx-tracing.jar

RUN	chown -R ${SERVICE_USER}:${SERVICE_GROUP} ./*.jar

USER ${SERVICE_USER}

# I added "-javaagent:signalfx-tracing.jar" to the JAVA_OPTS env variable in
# the CloudFormation template in the Fargate task definition, so we can more easily
# disable/enable tracing via that variable.
# Because we use the "exec" form with ENTRYPOINT, we do not need to add JAVA_OPTS to the
# command line -- in contrast to using the "shell" form that you will see in other Sock
# Shop services.
ENTRYPOINT ["/usr/local/bin/java.sh","-jar","./app.jar", "--port=80"]
