from uuid import uuid4

import boto3
import base64
import json
from aws_lambda_powertools.event_handler import (
    APIGatewayHttpResolver,
    Response,
    content_types,
)
from aws_lambda_powertools.utilities.typing import LambdaContext
from pydantic import BaseModel, Field
from pydantic_settings import BaseSettings
from aws_lambda_powertools import Logger

logger = Logger()


class Settings(BaseSettings):
    books_table_arn: str = ""


settings = Settings()

dynamodb = boto3.resource("dynamodb")
books_table = dynamodb.Table(settings.books_table_arn)

app = APIGatewayHttpResolver(enable_validation=True)


class BooksRecord(BaseModel):
    id: str = Field(alias="Id")
    title: str = Field(alias="Title")
    isbn: str = Field(alias="ISBN")
    authors: list[str] = Field(alias="Authors")
    pages: int = Field(alias="Pages")


class SaveBookRequest(BaseModel):
    title: str
    isbn: str
    authors: list[str]
    pages: int


class GetBookResponse(BaseModel):
    id: str
    title: str
    isbn: str
    authors: list[str]
    pages: int


class GetBooksResponse(BaseModel):
    books: list[GetBookResponse]
    last_evaluated_key: str | None


def to_get_book_response(item: dict) -> GetBookResponse:
    item = BooksRecord.model_validate(item)
    return GetBookResponse(
        id=item.id,
        title=item.title,
        isbn=item.isbn,
        authors=item.authors,
        pages=item.pages,
    )


@app.get("/books")
def get_books():
    params = app.current_event.query_string_parameters
    limit = params.get("limit")
    encoded = params.get("lastEvaluatedKey")
    params = {}
    if limit is not None:
        params["Limit"] = limit
    if encoded is not None:
        params["ExclusiveStartKey"] = json.loads(base64.urlsafe_b64decode(encoded))
    result = books_table.scan(**params)
    last_evaluated_key = result.get("LastEvaluatedKey")
    if last_evaluated_key is not None:
        encoded = base64.urlsafe_b64encode(
            json.dumps(last_evaluated_key).encode("utf-8")
        ).decode("utf-8")
    else:
        encoded = None
    books = list(map(lambda item: to_get_book_response(item), result["Items"]))
    return GetBooksResponse(
        books=books,
        last_evaluated_key=encoded,
    )


@app.get("/books/<book_id>")
def get_book(book_id: str) -> Response[GetBookResponse]:
    response = books_table.get_item(Key={"Id": book_id})
    item = response.get("Item")
    if item is None:
        return Response(status_code=404)
    else:
        return Response(
            status_code=200,
            content_type=content_types.APPLICATION_JSON,
            body=to_get_book_response(item),
        )


@app.post("/books")
def save_book(book: SaveBookRequest):
    book_id = str(uuid4())
    record = BooksRecord(
        Id=book_id,
        Title=book.title,
        ISBN=book.isbn,
        Authors=book.authors,
        Pages=book.pages,
    )
    item = record.model_dump(by_alias=True)
    logger.info(f"Saving: {item}")
    books_table.put_item(Item=item)
    return GetBookResponse(
        id=book_id,
        title=book.title,
        isbn=book.isbn,
        authors=book.authors,
        pages=book.pages,
    )


@logger.inject_lambda_context(log_event=True)
def handler(event: dict, context: LambdaContext) -> dict:
    return app.resolve(event, context)
