FROM maven:3.9.9-eclipse-temurin-21 AS builder
WORKDIR /src/web-node-java
COPY web-node-java/pom.xml ./
COPY web-node-java/src ./src
COPY proto /src/proto
RUN mvn -q -DskipTests package

FROM eclipse-temurin:21-jre-jammy
WORKDIR /app
COPY --from=builder /src/web-node-java/target/web-node-java-0.0.1.jar /app/app.jar
EXPOSE 8080
ENTRYPOINT ["java","-jar","/app/app.jar"]
