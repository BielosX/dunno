package org.dunno;

import org.dunno.dynamodb.BookEntity;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import software.amazon.awssdk.enhanced.dynamodb.DynamoDbEnhancedClient;
import software.amazon.awssdk.enhanced.dynamodb.DynamoDbTable;
import software.amazon.awssdk.enhanced.dynamodb.TableSchema;

@Configuration
public class AwsDynamoDbConfig {

  @Bean
  public DynamoDbEnhancedClient enhancedClient() {
    return DynamoDbEnhancedClient.create();
  }

  @Bean
  public DynamoDbTable<BookEntity> bookEntityTable(DynamoDbEnhancedClient client) {
    String books = System.getenv("AWS_DYNAMODB_TABLE_BOOKS");
    return client.table(books, TableSchema.fromBean(BookEntity.class));
  }
}
