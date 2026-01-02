package org.dunno

import com.amazonaws.services.lambda.runtime.Context
import com.amazonaws.services.lambda.runtime.RequestHandler
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyRequestEvent
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyResponseEvent
import org.http4k.routing.routes

class Handler : RequestHandler<APIGatewayProxyRequestEvent, APIGatewayProxyResponseEvent> {
  val rootRoutes = routes(statusRoutes)

  override fun handleRequest(
      input: APIGatewayProxyRequestEvent,
      context: Context?,
  ): APIGatewayProxyResponseEvent {
    val request = toRequest(input)
    val response = rootRoutes.match(request).invoke(request)
    return fromResponse(response)
  }
}
