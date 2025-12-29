package org.dunno;

import static org.dunno.ApiGwMapper.getLambdaContext;
import static org.springframework.web.servlet.function.RequestPredicates.contentType;
import static org.springframework.web.servlet.function.RouterFunctions.route;

import gg.jte.TemplateEngine;
import gg.jte.output.StringOutput;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;
import org.dunno.dynamodb.BookEntity;
import org.dunno.model.BookSaved;
import org.dunno.model.Books;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.stereotype.Service;
import org.springframework.util.MultiValueMap;
import org.springframework.web.servlet.function.HandlerFunction;
import org.springframework.web.servlet.function.RouterFunction;
import org.springframework.web.servlet.function.ServerRequest;
import org.springframework.web.servlet.function.ServerResponse;
import software.amazon.awssdk.enhanced.dynamodb.DynamoDbTable;
import tools.jackson.databind.ObjectMapper;

@Service
public class BooksResource {
  private static final Logger logger = LoggerFactory.getLogger(BooksResource.class);
  private static final String BOOKS_TEMPLATE = "books.jte";
  private static final String BOOK_SAVED_TEMPLATE = "bookSaved.jte";

  private final DynamoDbTable<BookEntity> booksTable;
  private final TemplateEngine engine;
  private final ObjectMapper objectMapper = new ObjectMapper();
  private final RouterFunction<ServerResponse> routes =
      route()
          .GET("/books", this::getBooks)
          .POST("/books", contentType(MediaType.APPLICATION_FORM_URLENCODED), this::saveBook)
          .build();

  public BooksResource(TemplateEngine engine, DynamoDbTable<BookEntity> table) {
    this.engine = engine;
    this.booksTable = table;
  }

  public ServerResponse getBooks(ServerRequest request) {
    String requestId = getLambdaContext(request).getAwsRequestId();
    List<Books.Book> books =
        booksTable.scan().items().stream()
            .map(i -> new Books.Book(i.getId(), i.getTitle(), i.getAuthors(), i.getReleaseDate()))
            .toList();
    Books model = new Books(books, requestId);
    StringOutput output = new StringOutput();
    engine.render(BOOKS_TEMPLATE, model, output);
    return ServerResponse.ok().contentType(MediaType.TEXT_HTML).body(output.toString());
  }

  public ServerResponse saveBook(ServerRequest request) {
    try {
      MultiValueMap<String, String> params = request.params();
      logger.info("Request params: {}", params);
      Map<String, String> body =
          params.entrySet().stream()
              .collect(Collectors.toMap(Map.Entry::getKey, e -> e.getValue().getFirst()));
      BookEntity entity = objectMapper.convertValue(body, BookEntity.class);
      UUID bookId = UUID.randomUUID();
      entity.setId(bookId);
      booksTable.putItem(entity);
      BookSaved model = new BookSaved(bookId, entity.getTitle());
      StringOutput output = new StringOutput();
      engine.render(BOOK_SAVED_TEMPLATE, model, output);
      return ServerResponse.ok().contentType(MediaType.TEXT_HTML).body(output.toString());
    } catch (Exception e) {
      logger.error(e.getMessage());
      return ServerResponse.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
    }
  }

  public Optional<HandlerFunction<ServerResponse>> getRoute(ServerRequest request) {
    return routes.route(request);
  }
}
