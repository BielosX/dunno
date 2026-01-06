package org.dunno

import com.squareup.moshi.JsonAdapter
import com.squareup.moshi.Moshi
import com.squareup.moshi.kotlin.reflect.KotlinJsonAdapterFactory
import software.amazon.awssdk.services.ssm.model.GetParameterRequest

data class DbConfig(
    val host: String,
    val username: String,
    val port: Int,
    val name: String,
    val maxPoolSize: Int,
    val minIdle: Int,
    val sslMode: String,
)

data class Config(val db: DbConfig)

val moshi: Moshi = Moshi.Builder().addLast(KotlinJsonAdapterFactory()).build()
val jsonAdapter: JsonAdapter<Config> = moshi.adapter(Config::class.java)

fun loadConfigFromSsm(key: String): Config {
  val request = GetParameterRequest.builder().name(key).build()
  val value = ssmClient.getParameter(request).parameter().value()
  return jsonAdapter.fromJson(value)!!
}

// Env Variables are not available during Lambda Managed Instance startup :(
val config = loadConfigFromSsm("KotlinConnectionPoolConfig")
