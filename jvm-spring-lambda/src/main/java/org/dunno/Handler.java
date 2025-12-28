package org.dunno;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.RequestHandler;
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyRequestEvent;
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyResponseEvent;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import org.jspecify.annotations.NullMarked;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.ApplicationContext;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.converter.HttpMessageConverter;
import org.springframework.http.converter.StringHttpMessageConverter;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.mock.web.MockHttpServletResponse;
import org.springframework.web.servlet.function.ServerRequest;
import org.springframework.web.servlet.function.ServerResponse;

@SpringBootApplication
public class Handler
    implements RequestHandler<APIGatewayProxyRequestEvent, APIGatewayProxyResponseEvent> {
  private static final Logger logger = LoggerFactory.getLogger(Handler.class);
  private static final ApplicationContext applicationContext = SpringApplication.run(Handler.class);

  private static ServerRequest toServerRequest(APIGatewayProxyRequestEvent event) {
    MockHttpServletRequest request = new MockHttpServletRequest();
    request.setRequestURI(event.getPath());
    request.setMethod(event.getHttpMethod());
    request.setParameters(event.getPathParameters());
    event.getHeaders().forEach(request::addHeader);
    event.getHeaders().entrySet().stream()
        .filter(e -> e.getKey().equalsIgnoreCase(HttpHeaders.CONTENT_TYPE))
        .findFirst()
        .map(Map.Entry::getValue)
        .ifPresent(request::setContentType);
    return ServerRequest.create(request, List.of(new StringHttpMessageConverter()));
  }

  private static class ResponseContext implements ServerResponse.Context {

    @Override
    @NullMarked
    public List<HttpMessageConverter<?>> messageConverters() {
      return List.of(new StringHttpMessageConverter());
    }
  }

  private static APIGatewayProxyResponseEvent fromServerResponse(ServerResponse response) {
    MockHttpServletRequest mockRequest = new MockHttpServletRequest();
    MockHttpServletResponse mockResponse = new MockHttpServletResponse();
    APIGatewayProxyResponseEvent event = new APIGatewayProxyResponseEvent();
    try {
      response.writeTo(mockRequest, mockResponse, new ResponseContext());
      event.setStatusCode(mockResponse.getStatus());
      event.setBody(mockResponse.getContentAsString());
      Map<String, String> headers = new HashMap<>();
      mockResponse.getHeaderNames().forEach(n -> headers.put(n, mockResponse.getHeader(n)));
      event.setHeaders(headers);
      event.setIsBase64Encoded(false);
    } catch (Exception e) {
      logger.error(e.getMessage());
      event.setStatusCode(HttpStatus.INTERNAL_SERVER_ERROR.value());
      return event;
    }
    return event;
  }

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
    logger.info("Received {} request, path: {}", input.getHttpMethod(), input.getPath());
    ServerRequest request = toServerRequest(input);
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
