import {
  CloudWatchClient,
  PutMetricDataCommand,
  PutMetricDataInput,
} from "@aws-sdk/client-cloudwatch";
import {
  SQSClient,
  SendMessageCommandInput,
  SendMessageCommand,
} from "@aws-sdk/client-sqs";
import {
  AttributeValue,
  DeleteItemCommand,
  DeleteItemCommandInput,
  DynamoDBClient,
  PutItemCommand,
  PutItemCommandInput,
} from "@aws-sdk/client-dynamodb";
import { marshall } from "@aws-sdk/util-dynamodb";

export const putCloudWatchMetric = async (
  client: CloudWatchClient,
  namespace: string,
  metricName: string,
  value: number,
  timestamp: Date,
) => {
  const input: PutMetricDataInput = {
    Namespace: namespace,
    MetricData: [
      {
        MetricName: metricName,
        Value: value,
        Timestamp: timestamp,
      },
    ],
  };
  const command = new PutMetricDataCommand(input);
  await client.send(command);
};

export const sendSqsMessage = async (
  client: SQSClient,
  body: string,
  queueUrl: string,
) => {
  const input: SendMessageCommandInput = {
    MessageBody: body,
    QueueUrl: queueUrl,
  };
  const command = new SendMessageCommand(input);
  await client.send(command);
};

export const putDynamoDbItem = async <T>(
  client: DynamoDBClient,
  table: string,
  data: T,
) => {
  const item: Record<string, AttributeValue> = marshall(data, {
    removeUndefinedValues: true,
    convertEmptyValues: true,
  });
  const input: PutItemCommandInput = {
    Item: item,
    TableName: table,
  };
  const command = new PutItemCommand(input);
  await client.send(command);
};

export const deleteDynamoDbItem = async <K>(
  client: DynamoDBClient,
  table: string,
  key: K,
) => {
  const itemKey: Record<string, AttributeValue> = marshall(key);
  const input: DeleteItemCommandInput = {
    Key: itemKey,
    TableName: table,
  };
  const command = new DeleteItemCommand(input);
  await client.send(command);
};
