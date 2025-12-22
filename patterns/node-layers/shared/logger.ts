import { destination, pino } from "pino";

export const logger = pino(
  destination({
    sync: false,
  }),
);
