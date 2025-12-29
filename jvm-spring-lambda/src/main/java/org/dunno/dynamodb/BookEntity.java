package org.dunno.dynamodb;

import java.time.LocalDate;
import java.util.UUID;
import lombok.Data;
import lombok.Setter;
import software.amazon.awssdk.enhanced.dynamodb.mapper.annotations.DynamoDbBean;
import software.amazon.awssdk.enhanced.dynamodb.mapper.annotations.DynamoDbPartitionKey;

@Data
@DynamoDbBean
public class BookEntity {
  @Setter(onMethod_ = @DynamoDbPartitionKey)
  private UUID id;

  private String title;
  private String authors;
  private LocalDate releaseDate;
}
