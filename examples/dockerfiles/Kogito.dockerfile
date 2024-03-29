FROM quay.io/kiegroup/kogito-swf-builder-nightly:latest AS builder

# Kogito User
USER 1001

ARG QUARKUS_PACKAGE_TYPE="jar"
ARG SCRIPT_DEBUG="false"

 # Copy from build context to skeleton resources project
COPY * ./resources/

RUN /home/kogito/launch/build-app.sh ./resources
#=============================
# Runtime Run
#=============================
FROM registry.access.redhat.com/ubi8/openjdk-11-runtime:latest

ARG QUARKUS_LAUNCH_DEVMODE=false

ENV LANG='en_US.UTF-8' LANGUAGE='en_US:en'
# We make four distinct layers so if there are application changes the library layers can be re-used
COPY --from=builder --chown=185 /home/kogito/serverless-workflow-project/target/quarkus-app/lib/ /deployments/lib/
COPY --from=builder --chown=185 /home/kogito/serverless-workflow-project/target/quarkus-app/*.jar /deployments/
COPY --from=builder --chown=185 /home/kogito/serverless-workflow-project/target/quarkus-app/app/ /deployments/app/
COPY --from=builder --chown=185 /home/kogito/serverless-workflow-project/target/quarkus-app/quarkus/ /deployments/quarkus/

EXPOSE 8080
USER 185
ENV AB_JOLOKIA_OFF=""
ENV JAVA_OPTS="-Dquarkus.http.host=0.0.0.0 -Djava.util.logging.manager=org.jboss.logmanager.LogManager"
ENV JAVA_APP_JAR="/deployments/quarkus-run.jar"
