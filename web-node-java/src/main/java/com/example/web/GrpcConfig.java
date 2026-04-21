package com.example.web;

import com.example.social.v1.PostServiceGrpc;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class GrpcConfig {
    @Bean
    public ManagedChannel managedChannel(@Value("${grpc.target:core-service:9090}") String target) {
        return ManagedChannelBuilder.forTarget(target).usePlaintext().build();
    }

    @Bean
    public PostServiceGrpc.PostServiceBlockingStub postStub(ManagedChannel channel) {
        return PostServiceGrpc.newBlockingStub(channel);
    }
}
