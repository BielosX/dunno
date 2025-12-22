import {
  GetObjectCommand,
  GetObjectCommandInput,
  PutObjectCommand,
  PutObjectCommandInput,
  S3Client,
} from "@aws-sdk/client-s3";
import JSZip from "jszip";
import sharp from "sharp";
import {
  GetFunctionConfigurationCommand,
  GetFunctionConfigurationCommandInput,
  LambdaClient,
} from "@aws-sdk/client-lambda";
import { parse } from "@aws-sdk/util-arn-parser";

export const uploadS3File = async (
  client: S3Client,
  bucket: string,
  key: string,
  body: Uint8Array,
) => {
  console.info(`Uploading: ${[bucket, key]}`);
  const input: PutObjectCommandInput = {
    Bucket: bucket,
    Key: key,
    Body: body,
    ContentLength: body.length,
  };
  const command = new PutObjectCommand(input);
  await client.send(command);
};

export const downloadS3File = async (
  client: S3Client,
  bucket: string,
  key: string,
): Promise<Uint8Array | undefined> => {
  console.info(`Downloading: ${[bucket, key]}`);
  const input: GetObjectCommandInput = {
    Bucket: bucket,
    Key: key,
  };
  const command = new GetObjectCommand(input);
  const result = await client.send(command);
  if (!result.Body) {
    return undefined;
  }
  return await result.Body.transformToByteArray();
};

export const zipS3File = async (
  client: S3Client,
  sourceBucket: string,
  sourceKey: string,
  targetBucket: string,
) => {
  console.info(
    `zipS3File called with params: ${[sourceBucket, sourceKey, targetBucket]}`,
  );
  try {
    const file = await downloadS3File(client, sourceBucket, sourceKey);
    if (file !== undefined) {
      console.info(`File of size ${file.length} downloaded`);
      let zip = new JSZip();
      zip.file(sourceKey, file);
      const buffer = await zip.generateAsync({ type: "uint8array" });
      console.info(`Uploading file of size ${buffer.length}`);
      await uploadS3File(client, targetBucket, `${sourceKey}.zip`, buffer);
      console.info("Uploaded zip file");
    } else {
      console.warn(`File ${sourceKey} not found.`);
    }
  } catch (e) {
    console.error(e);
  }
};

export const getUsedLayers = async (client: LambdaClient, arn: string) => {
  const input: GetFunctionConfigurationCommandInput = {
    FunctionName: arn,
  };
  const command = new GetFunctionConfigurationCommand(input);
  const response = await client.send(command);
  const layers = response.Layers || [];
  return layers.map((layer) => {
    const layerArn = layer.Arn as string;
    return parse(layerArn).resource;
  });
};

export const resizeS3File = async (
  client: S3Client,
  sourceBucket: string,
  sourceKey: string,
  targetBucket: string,
) => {
  console.info(
    `resizeS3File called with params: ${[sourceBucket, sourceKey, targetBucket]}`,
  );
  try {
    const file = await downloadS3File(client, sourceBucket, sourceKey);
    if (file !== undefined) {
      const s = sharp(file);
      const meta = await s.metadata();
      console.info(`Resizing file ${sourceKey} with format ${meta.format}`);
      const buffer: Buffer = await s
        .resize(512, 512)
        .toBuffer({ resolveWithObject: false });
      const array: Uint8Array = new Uint8Array(buffer);
      await uploadS3File(client, targetBucket, sourceKey, array);
      console.info("Uploaded image file");
    } else {
      console.warn(`File ${sourceKey} not found.`);
    }
  } catch (e) {
    console.error(e);
  }
  return;
};
