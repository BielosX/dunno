import { Context } from "aws-lambda";
import { randomUUID } from "crypto";

type Event = {
  sleepSeconds: number;
};

type Response = {
  logStream: string;
  envId: string;
  requestId: string;
};

let envId: string | undefined = undefined;

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

export const handler = async (
  event: Event,
  ctx: Context,
): Promise<Response> => {
  const logStream = ctx.logStreamName;
  if (envId == undefined) {
    envId = randomUUID();
  }
  await sleep(event.sleepSeconds * 1000);
  return {
    logStream,
    envId,
    requestId: ctx.awsRequestId,
  };
};
