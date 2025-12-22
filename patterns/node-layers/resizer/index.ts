import { Context, S3Event } from "aws-lambda";
import { z } from "zod";
import { S3Client } from "@aws-sdk/client-s3";
import { LambdaClient } from "@aws-sdk/client-lambda";
import { getUsedLayers, resizeS3File } from "@shared/services";
import { logger } from "@shared/logger";

const EnvSchema = z.object({
  TARGET_BUCKET: z.string(),
});

const s3Client = new S3Client();
const lambdaClient = new LambdaClient();

export const resize = async (event: S3Event, ctx: Context) => {
  const env = EnvSchema.parse(process.env);
  const layers = await getUsedLayers(lambdaClient, ctx.invokedFunctionArn);
  logger.info(`Used layers: ${JSON.stringify(layers)}`);
  await Promise.all(
    event.Records.map((item) => {
      const key = item.s3.object.key;
      const bucket = item.s3.bucket.name;
      logger.info(`Trying to resize file ${key}`);
      return resizeS3File(s3Client, bucket, key, env.TARGET_BUCKET);
    }),
  );
  logger.flush();
};
