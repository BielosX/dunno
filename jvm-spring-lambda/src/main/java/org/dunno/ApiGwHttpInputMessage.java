package org.dunno;

import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyRequestEvent;
import java.io.ByteArrayInputStream;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.util.Base64;
import org.jspecify.annotations.NullMarked;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpInputMessage;
import org.springframework.util.LinkedMultiValueMap;
import org.springframework.util.MultiValueMap;

public class ApiGwHttpInputMessage implements HttpInputMessage {
  private final String body;
  private final HttpHeaders headers;

  public ApiGwHttpInputMessage(APIGatewayProxyRequestEvent request) {
    if (request.getIsBase64Encoded()) {
      byte[] decodedBytes = Base64.getDecoder().decode(request.getBody());
      this.body = new String(decodedBytes);
    } else {
      this.body = request.getBody();
    }
    MultiValueMap<String, String> multiMap = new LinkedMultiValueMap<>();
    multiMap.putAll(request.getMultiValueHeaders());
    this.headers = HttpHeaders.copyOf(multiMap);
  }

  @Override
  @NullMarked
  public InputStream getBody() {
    return new ByteArrayInputStream(this.body.getBytes(StandardCharsets.UTF_8));
  }

  @Override
  @NullMarked
  public HttpHeaders getHeaders() {
    return this.headers;
  }
}
