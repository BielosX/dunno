package org.dunno

import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyRequestEvent
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyResponseEvent
import kotlin.io.encoding.Base64
import org.http4k.core.Method
import org.http4k.core.Request
import org.http4k.core.Response

fun toRequest(event: APIGatewayProxyRequestEvent): Request {
  val request = Request(Method.valueOf(event.httpMethod), event.path)
  if (event.body != null) {
    request.body(event.body)
    if (event.isBase64Encoded) {
      val content = Base64.decode(event.body).decodeToString()
      request.body(content)
    }
  } else {
    request.body(String())
  }
  val headers = event.headers.map { entry -> Pair(entry.key, entry.value) }.toList()
  request.headers(headers)
  return request
}

fun fromResponse(response: Response): APIGatewayProxyResponseEvent {
  val event = APIGatewayProxyResponseEvent()
  event.body = response.body.text
  event.statusCode = response.status.code
  event.isBase64Encoded = false
  event.headers = response.headers.toMap()
  return event
}
