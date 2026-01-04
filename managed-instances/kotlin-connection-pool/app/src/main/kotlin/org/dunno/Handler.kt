package org.dunno

import com.amazonaws.services.lambda.runtime.Context
import com.amazonaws.services.lambda.runtime.RequestHandler
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyRequestEvent
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyResponseEvent
import org.http4k.routing.routes
import org.jetbrains.exposed.sql.Database

class Handler : RequestHandler<APIGatewayProxyRequestEvent, APIGatewayProxyResponseEvent> {
  val rootRoutes = routes(statusRoutes)

  init {
    Database.connect(dataSource)
  }

  override fun handleRequest(
      input: APIGatewayProxyRequestEvent,
      context: Context?,
  ): APIGatewayProxyResponseEvent {
    val request = toRequest(input)
    val response = rootRoutes.match(request).invoke(request)
    return fromResponse(response)
  }
}
