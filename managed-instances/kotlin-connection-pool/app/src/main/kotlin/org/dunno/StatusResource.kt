package org.dunno

import org.http4k.core.Method
import org.http4k.core.Request
import org.http4k.core.Response
import org.http4k.core.Status
import org.http4k.routing.bind
import org.http4k.routing.routes

val getStatus = { _: Request -> Response(Status.OK) }

val statusRoutes =
    routes(
        "/status" bind Method.GET to getStatus,
    )
