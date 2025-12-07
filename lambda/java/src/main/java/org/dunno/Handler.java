package org.dunno;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.RequestHandler;

@SuppressWarnings("unused")
public class Handler implements RequestHandler<Void, String> {
  @Override
  public String handleRequest(Void unused, Context context) {
    System.out.printf(
        "AwsRequestId: %s, FunctionArn: %s",
        context.getAwsRequestId(), context.getInvokedFunctionArn());
    return "OK";
  }
}
