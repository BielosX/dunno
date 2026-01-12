use lambda_http::{run, service_fn, tracing, Error};
mod http_handler;
use crate::http_handler::get_books;
use axum::{routing::get, Router};
use http_handler::function_handler;

#[tokio::main]
async fn main() -> Result<(), Error> {
    tracing::init_default_subscriber();
    let router = Router::new().route("/books", get(get_books));

    run(service_fn(move |req| {
        let router = router.clone();
        async move { function_handler(router, req).await }
    }))
    .await
}
