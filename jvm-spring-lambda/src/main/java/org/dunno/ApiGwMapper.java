package org.dunno;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyRequestEvent;
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyResponseEvent;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import lombok.SneakyThrows;
import org.jspecify.annotations.NullMarked;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.converter.FormHttpMessageConverter;
import org.springframework.http.converter.HttpMessageConverter;
import org.springframework.http.converter.StringHttpMessageConverter;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.mock.web.MockHttpServletResponse;
import org.springframework.util.MultiValueMap;
import org.springframework.web.servlet.function.ServerRequest;
import org.springframework.web.servlet.function.ServerResponse;

public class ApiGwMapper {
  private static final String LAMBDA_CONTEXT = "LambdaContext";
  private static final Logger logger = LoggerFactory.getLogger(ApiGwMapper.class);
  private static final FormHttpMessageConverter converter = new FormHttpMessageConverter();

  public static Context getLambdaContext(ServerRequest request) {
    return (Context) request.attribute(LAMBDA_CONTEXT).orElseThrow();
  }

  @SneakyThrows
  public static ServerRequest toServerRequest(APIGatewayProxyRequestEvent event, Context context) {
    MockHttpServletRequest request = new MockHttpServletRequest();
    request.setRequestURI(event.getPath());
    request.setMethod(event.getHttpMethod());
    request.setAttribute(LAMBDA_CONTEXT, context);
    if (event.getBody() != null) {
      MultiValueMap<String, String> message =
          converter.read(null, new ApiGwHttpInputMessage(event));
      message.forEach((key, values) -> request.setParameter(key, values.toArray(String[]::new)));
    }
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

  public static APIGatewayProxyResponseEvent fromServerResponse(ServerResponse response) {
    MockHttpServletRequest mockRequest = new MockHttpServletRequest();
    MockHttpServletResponse mockResponse = new MockHttpServletResponse();
    APIGatewayProxyResponseEvent event = new APIGatewayProxyResponseEvent();
    try {
      response.writeTo(mockRequest, mockResponse, new ApiGwMapper.ResponseContext());
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
}
