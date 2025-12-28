package org.dunno;

import static org.springframework.web.servlet.function.RouterFunctions.route;

import java.util.Optional;
import org.springframework.stereotype.Service;
import org.springframework.web.servlet.function.HandlerFunction;
import org.springframework.web.servlet.function.RouterFunction;
import org.springframework.web.servlet.function.ServerRequest;
import org.springframework.web.servlet.function.ServerResponse;

@Service
public class BooksResource {
  private final RouterFunction<ServerResponse> routes =
      route().GET("/books", this::getBooks).build();

  public ServerResponse getBooks(ServerRequest request) {
    return ServerResponse.ok().body("Hello");
  }

  public Optional<HandlerFunction<ServerResponse>> getRoute(ServerRequest request) {
    return routes.route(request);
  }
}
