package org.dunno

import com.zaxxer.hikari.HikariConfig
import com.zaxxer.hikari.HikariCredentialsProvider
import com.zaxxer.hikari.HikariDataSource
import com.zaxxer.hikari.util.Credentials
import software.amazon.awssdk.auth.credentials.EnvironmentVariableCredentialsProvider
import software.amazon.awssdk.services.rds.model.GenerateAuthenticationTokenRequest

class AwsCredentialsProvider : HikariCredentialsProvider {
  fun getToken(): String {
    val utils = rdsClient.utilities()
    val tokenRequest =
        GenerateAuthenticationTokenRequest.builder()
            .credentialsProvider(EnvironmentVariableCredentialsProvider.create())
            .username(config.db.username)
            .port(config.db.port)
            .hostname(config.db.host)
            .build()
    return utils.generateAuthenticationToken(tokenRequest)
  }

  override fun getCredentials(): Credentials? {
    val token = getToken()
    return Credentials.of(config.db.username, token)
  }
}

val cfg =
    HikariConfig().apply {
      maximumPoolSize = 10
      minimumIdle = 2
      username = config.db.username
      credentialsProvider = AwsCredentialsProvider()
      jdbcUrl =
          "jdbc:postgresql://${config.db.host}:${config.db.port}/${config.db.name}?sslmode=require&tcpKeepAlive=true"
    }
val dataSource = HikariDataSource(cfg)
