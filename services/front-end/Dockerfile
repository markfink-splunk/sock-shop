FROM node:10-alpine
ENV NODE_ENV "production"
ENV PORT 8079
EXPOSE 8079
RUN addgroup mygroup && adduser -D -G mygroup myuser && mkdir -p /usr/src/app && chown -R myuser /usr/src/app

# Prepare app directory
WORKDIR /usr/src/app
COPY package.json /usr/src/app/
COPY yarn.lock /usr/src/app/
RUN chown myuser /usr/src/app/yarn.lock

USER myuser
RUN yarn install

COPY . /usr/src/app

# SIGNALFX UAPM AUTO-INSTRUMENTATION
COPY signalfx-nodejs-tracing-0.8.0.tar.gz /usr/src/app
RUN tar xzf signalfx-nodejs-tracing-0.8.0.tar.gz \
    && npm install ./signalfx-nodejs-tracing-0.8.0

# Start the app
CMD ["/usr/local/bin/npm", "start"]
