use actix_web::{HttpResponse, ResponseError};
use derive_more::Display;
use validator::ValidationErrors;
use diesel::result::Error as DieselError;

#[derive(Debug, Display)]
pub enum ServiceError {
    #[display(fmt = "Internal Server Error")]
    InternalServerError,

    #[display(fmt = "BadRequest: {}", _0)]
    BadRequest(String),

    #[display(fmt = "Unauthorized: {}", _0)]
    Unauthorized(String),

    #[display(fmt = "NotFound: {}", _0)]
    NotFound(String),

    #[display(fmt = "DatabaseError: {}", _0)]
    DatabaseError(String),
}

impl ResponseError for ServiceError {
    fn error_response(&self) -> HttpResponse {
        match *self {
            ServiceError::InternalServerError => {
                HttpResponse::InternalServerError().json("Internal Server Error")
            }
            ServiceError::BadRequest(ref message) => {
                HttpResponse::BadRequest().json(message)
            }
            ServiceError::Unauthorized(ref message) => {
                HttpResponse::Unauthorized().json(message)
            }
            ServiceError::NotFound(ref message) => {
                HttpResponse::NotFound().json(message)
            }
            ServiceError::DatabaseError(ref message) => {
                HttpResponse::InternalServerError().json(message)
            }
        }
    }
}

impl From<DieselError> for ServiceError {
    fn from(err: DieselError) -> ServiceError {
        match err {
            DieselError::NotFound => {
                ServiceError::NotFound("Record not found".to_string())
            }
            _ => ServiceError::DatabaseError(err.to_string()),
        }
    }
}

impl From<ValidationErrors> for ServiceError {
    fn from(err: ValidationErrors) -> ServiceError {
        ServiceError::BadRequest(err.to_string())
    }
}