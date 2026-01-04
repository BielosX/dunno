package org.dunno

import software.amazon.awssdk.auth.credentials.EnvironmentVariableCredentialsProvider
import software.amazon.awssdk.services.rds.RdsClient
import software.amazon.awssdk.services.ssm.SsmClient

val awsCredentialsProvider: EnvironmentVariableCredentialsProvider =
    EnvironmentVariableCredentialsProvider.create()
val rdsClient: RdsClient = RdsClient.builder().credentialsProvider(awsCredentialsProvider).build()
val ssmClient: SsmClient = SsmClient.builder().credentialsProvider(awsCredentialsProvider).build()
