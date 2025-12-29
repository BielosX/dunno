package org.dunno;

import static org.dunno.ApiGwMapper.fromServerResponse;
import static org.dunno.ApiGwMapper.toServerRequest;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.RequestHandler;
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyRequestEvent;
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyResponseEvent;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.context.properties.ConfigurationPropertiesScan;
import org.springframework.context.ApplicationContext;
import org.springframework.http.HttpStatus;
import org.springframework.web.servlet.function.ServerRequest;
import org.springframework.web.servlet.function.ServerResponse;

@SpringBootApplication
@ConfigurationPropertiesScan
public class Handler
    implements RequestHandler<APIGatewayProxyRequestEvent, APIGatewayProxyResponseEvent> {
  private static final Logger logger = LoggerFactory.getLogger(Handler.class);
  private static final ApplicationContext applicationContext = SpringApplication.run(Handler.class);

  private static APIGatewayProxyResponseEvent internalServerError() {
    APIGatewayProxyResponseEvent event = new APIGatewayProxyResponseEvent();
    event.setStatusCode(HttpStatus.INTERNAL_SERVER_ERROR.value());
    return event;
  }

  private static APIGatewayProxyResponseEvent notFound() {
    APIGatewayProxyResponseEvent event = new APIGatewayProxyResponseEvent();
    event.setStatusCode(HttpStatus.NOT_FOUND.value());
    return event;
  }

  @Override
  public APIGatewayProxyResponseEvent handleRequest(
      APIGatewayProxyRequestEvent input, Context context) {
    logger.info(
        "Received {} request, path: {}, body: {}",
        input.getHttpMethod(),
        input.getPath(),
        input.getBody());
    ServerRequest request = toServerRequest(input, context);
    BooksResource booksResource = applicationContext.getBean(BooksResource.class);
    return booksResource
        .getRoute(request)
        .map(
            f -> {
              try {
                ServerResponse response = f.handle(request);
                return fromServerResponse(response);
              } catch (Exception e) {
                logger.error(e.getMessage());
                return internalServerError();
              }
            })
        .orElse(notFound());
  }
}
