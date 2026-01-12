use axum::body::Body as AxumBody;
use axum::Router;
use http::Request as HttpRequest;
use lambda_http::{Body as LambdaBody, Body};
use lambda_http::{Error, Request, Response, Service};

pub async fn get_books() -> &'static str {
    "Hello, World!"
}

fn lambda_body_to_axum(body: &LambdaBody) -> AxumBody {
    match body {
        LambdaBody::Empty => AxumBody::empty(),
        LambdaBody::Text(text) => AxumBody::from(text.clone()),
        LambdaBody::Binary(bytes) => AxumBody::from(bytes.clone()),
        _ => AxumBody::empty(),
    }
}

async fn axum_to_lambda_body(body: AxumBody) -> Result<LambdaBody, axum_core::Error> {
    let bytes = axum::body::to_bytes(body, usize::MAX).await?;
    if bytes.is_empty() {
        Ok(LambdaBody::Empty)
    } else {
        Ok(LambdaBody::Text(
            String::from_utf8_lossy(&bytes).to_string(),
        ))
    }
}

pub async fn function_handler(router: Router, event: Request) -> Result<Response<Body>, Error> {
    let body = lambda_body_to_axum(event.body());
    let request = HttpRequest::builder().uri(event.uri()).body(body)?;

    let response = router.clone().as_service().call(request).await?;
    let status = response.status().clone();
    let response_body = axum_to_lambda_body(response.into_body()).await?;

    let resp = Response::builder()
        .status(status)
        .header("Content-Type", "application/json")
        .body(response_body)
        .map_err(Box::new)?;
    Ok(resp)
}
