def handler(_event, context):
    print(f"AwsRequestId: {context.aws_request_id}, InvokedFunctionArn: {context.invoked_function_arn}")
    return "OK"