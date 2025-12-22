import { Context, S3Event } from "aws-lambda";
import { z } from "zod";
import { S3Client } from "@aws-sdk/client-s3";
import { LambdaClient } from "@aws-sdk/client-lambda";
import { getUsedLayers, resizeS3File } from "@shared/services";

const EnvSchema = z.object({
  TARGET_BUCKET: z.string(),
});

const s3Client = new S3Client();
const lambdaClient = new LambdaClient();

export const resize = async (event: S3Event, ctx: Context) => {
  const env = EnvSchema.parse(process.env);
  const layers = await getUsedLayers(lambdaClient, ctx.invokedFunctionArn);
  console.info(`Used layers: ${JSON.stringify(layers)}`);
  for (const record of event.Records) {
    const key = record.s3.object.key;
    console.info(`Trying to resize file ${key}`);
    await resizeS3File(s3Client, record.s3.bucket.name, key, env.TARGET_BUCKET);
  }
};
