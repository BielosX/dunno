import { Context, KinesisStreamEvent, KinesisStreamRecord } from "aws-lambda";
import { pino, destination } from "pino";
import pLimit from "p-limit";
import { z } from "zod";
import { CloudWatchClient } from "@aws-sdk/client-cloudwatch";
import { SQSClient } from "@aws-sdk/client-sqs";
import { DynamoDBClient } from "@aws-sdk/client-dynamodb";
import * as protobuf from "protobufjs";
import {
  deleteDynamoDbItem,
  putCloudWatchMetric,
  putDynamoDbItem,
  sendSqsMessage,
} from "./aws.js";

enum UserEventType {
  Created = 0,
  Updated = 1,
  Deleted = 2,
}

type UserEventRecord = {
  eventType: UserEventType;
  eventId: string;
  userId: string;
  userName: string;
  userEmail: string;
  userTimezone: string;
};

type UserDynamoDbRecord = {
  userId: string;
  userName: string;
  userEmail: string;
  userTimezone: string;
  created?: string;
  updated?: string;
};

const logger = pino(
  destination({
    sync: true,
  }),
);

const envSchema = z.object({
  CONCURRENCY_LIMIT: z.coerce.number().int().positive().default(10),
  FAILED_RECORDS_QUEUE: z.string(),
  USERS_TABLE: z.string(),
  LOG_LEVEL: z.string().default("info"),
});

const cloudWatchClient = new CloudWatchClient();
const sqsClient = new SQSClient();
const dynamoDbClient = new DynamoDBClient();

export const handler = async (event: KinesisStreamEvent, ctx: Context) => {
  const root = await protobuf.load("protobuf/user_event.proto");
  const UserEvent = root.lookupType("users.UserEvent");
  const env = envSchema.parse(process.env);
  const sendFailedRecord = async (record: KinesisStreamRecord) => {
    await sendSqsMessage(
      sqsClient,
      JSON.stringify(record),
      env.FAILED_RECORDS_QUEUE,
    );
  };
  logger.level = env.LOG_LEVEL;
  logger.info(
    `Function ${ctx.invokedFunctionArn} started with Concurrency Limit ${env.CONCURRENCY_LIMIT}`,
  );
  const limit = pLimit(env.CONCURRENCY_LIMIT);
  let failedRecords = 0;
  await Promise.all(
    event.Records.map((record) =>
      limit(async () => {
        const data = Buffer.from(record.kinesis.data, "base64");
        try {
          const message = UserEvent.decode(data);
          const object = UserEvent.toObject(message, {
            enums: Number,
          }) as UserEventRecord;
          let user: UserDynamoDbRecord = {
            userEmail: object.userEmail,
            userId: object.userId,
            userName: object.userName,
            userTimezone: object.userTimezone,
          };
          logger.info(`Decoded message: ${JSON.stringify(object)}`);
          const now = new Date().toISOString();
          switch (object.eventType) {
            case UserEventType.Created:
              user.created = now;
              await putDynamoDbItem(dynamoDbClient, env.USERS_TABLE, user);
              logger.info(`User ${object.userName} created`);
              break;
            case UserEventType.Updated:
              user.updated = now;
              await putDynamoDbItem(dynamoDbClient, env.USERS_TABLE, user);
              logger.info(`User ${object.userName} updated`);
              break;
            case UserEventType.Deleted:
              logger.info(`User ${object.userName} deleted`);
              await deleteDynamoDbItem(dynamoDbClient, env.USERS_TABLE, {
                userId: object.userName,
              });
              break;
            default:
              logger.error(`Unexpected user event type for ${object.userName}`);
              failedRecords++;
              await sendFailedRecord(record);
              break;
          }
        } catch (e) {
          logger.error(
            `Unable to process a record ${record.eventID}, error: ${e}`,
          );
          failedRecords++;
          await sendFailedRecord(record);
        }
      }),
    ),
  );
  const now = new Date();
  await putCloudWatchMetric(
    cloudWatchClient,
    "UserEvents",
    "FailedKinesisRecords",
    failedRecords,
    now,
  );
};
