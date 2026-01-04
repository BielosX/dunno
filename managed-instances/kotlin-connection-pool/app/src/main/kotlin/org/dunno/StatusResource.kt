package org.dunno

import org.http4k.core.Body
import org.http4k.core.Method
import org.http4k.core.Request
import org.http4k.core.Response
import org.http4k.core.Status
import org.http4k.core.with
import org.http4k.format.Moshi.auto
import org.http4k.lens.BiDiBodyLens
import org.http4k.routing.bind
import org.http4k.routing.routes
import org.jetbrains.exposed.sql.transactions.transaction

data class AppStatus(val dbVersion: String, val activeDbConnections: Int, val idleDbConfig: Int)

val appStatusLens: BiDiBodyLens<AppStatus> = Body.auto<AppStatus>().toLens()

val getStatus = { _: Request ->
  val version: String = transaction {
    exec("SELECT version()") { rs ->
      rs.next()
      rs.getString(1)
    }!!
  }
  val mxBean = DatabaseFactory.dataSource.hikariPoolMXBean
  val result = AppStatus(version, mxBean.activeConnections, mxBean.idleConnections)
  Response(Status.OK).with(appStatusLens of result)
}

val statusRoutes =
    routes(
        "/status" bind Method.GET to getStatus,
    )
