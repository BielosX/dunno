use lambda_runtime::{run, service_fn, tracing, Error, LambdaEvent};
use serde_json::Value;
use std::env;

async fn function_handler(event: LambdaEvent<Value>) -> Result<String, Error> {
    let log_stream = env::var("AWS_LAMBDA_LOG_STREAM_NAME")?;
    let request_id = event.context.request_id;
    let function_arn = event.context.invoked_function_arn;
    println!(
        "AwsRequestId: {}, InvokedFunctionArn: {}",
        request_id, function_arn
    );
    Ok(log_stream)
}

#[tokio::main]
async fn main() -> Result<(), Error> {
    tracing::init_default_subscriber();

    run(service_fn(function_handler)).await
}
