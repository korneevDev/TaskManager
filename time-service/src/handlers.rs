use actix_web::{web, HttpResponse, get, post, put, delete};
use serde::{Deserialize, Serialize};
use serde_json::json;
use uuid::Uuid;
use jsonwebtoken::{decode, DecodingKey, Validation, Algorithm};
use validator::Validate;

use crate::{
    models::{CreateTimeEntry, TimeEntryResponse}, 
    repository::TimeEntryRepository, 
    errors::ServiceError
};

#[derive(Debug, Serialize, Deserialize)]
struct Claims {
    sub: String,
    exp: usize,
}

pub fn time_entry_handlers(cfg: &mut web::ServiceConfig) {
    cfg.service(
        web::scope("/time-entries")
            .service(start_time_entry)
            .service(stop_time_entry)
            .service(list_time_entries)
            .service(list_time_entries_by_task)
            .service(get_active_time_entry)
            .service(delete_time_entry),
    );
}

#[get("/health")]
pub async fn health_check() -> HttpResponse {
    HttpResponse::Ok().json(json!({"status": "ok"}))
}

async fn extract_user_id(
    req: &actix_web::HttpRequest,
    jwt_secret: &str,
) -> Result<Uuid, ServiceError> {
    let auth_header = req.headers().get("Authorization")
        .ok_or(ServiceError::Unauthorized("Authorization header is missing".into()))?
        .to_str()
        .map_err(|_| ServiceError::Unauthorized("Invalid Authorization header".into()))?;

    let parts: Vec<&str> = auth_header.split_whitespace().collect();
    if parts.len() != 2 || parts[0] != "Bearer" {
        return Err(ServiceError::Unauthorized("Invalid token format".into()));
    }

    let token = parts[1];
    let token_data = decode::<Claims>(
        token,
        &DecodingKey::from_secret(jwt_secret.as_ref()),
        &Validation::new(Algorithm::HS256),
    ).map_err(|_| ServiceError::Unauthorized("Invalid token".into()))?;

    Uuid::parse_str(&token_data.claims.sub)
        .map_err(|_| ServiceError::Unauthorized("Invalid user ID in token".into()))
}

#[post("")]
pub async fn start_time_entry(
    req: actix_web::HttpRequest,
    repo: web::Data<TimeEntryRepository>,
    jwt_secret: web::Data<String>,
    entry: web::Json<CreateTimeEntry>,
) -> Result<HttpResponse, ServiceError> {
    entry.validate()?;
    let user_id = extract_user_id(&req, &jwt_secret).await?;
    
    let entry = repo.create(user_id, entry.into_inner())?;
    Ok(HttpResponse::Created().json(entry))
}

#[put("/{id}/stop")]
pub async fn stop_time_entry(
    req: actix_web::HttpRequest,
    repo: web::Data<TimeEntryRepository>,
    jwt_secret: web::Data<String>,
    id: web::Path<Uuid>,
) -> Result<HttpResponse, ServiceError> {
    let user_id = extract_user_id(&req, &jwt_secret).await?;
    
    let entry = repo.stop(id.into_inner(), user_id)?;
    Ok(HttpResponse::Ok().json(entry))
}


#[get("")]
pub async fn list_time_entries(
    req: actix_web::HttpRequest,
    repo: web::Data<TimeEntryRepository>,
    jwt_secret: web::Data<String>,
) -> Result<HttpResponse, ServiceError> {
    let user_id = extract_user_id(&req, &jwt_secret).await?;
    
    let entries = repo.list_by_user(user_id)?;
    Ok(HttpResponse::Ok().json(entries))
}

#[get("/task/{task_id}")]
pub async fn list_time_entries_by_task(
    req: actix_web::HttpRequest,
    repo: web::Data<TimeEntryRepository>,
    jwt_secret: web::Data<String>,
    task_id: web::Path<Uuid>,
) -> Result<HttpResponse, ServiceError> {
    let user_id = extract_user_id(&req, &jwt_secret).await?;
    
    let entries = repo.list_by_task(user_id, task_id.into_inner())?;
    Ok(HttpResponse::Ok().json(entries))
}

#[get("/active")]
pub async fn get_active_time_entry(
    req: actix_web::HttpRequest,
    repo: web::Data<TimeEntryRepository>,
    jwt_secret: web::Data<String>,
) -> Result<HttpResponse, ServiceError> {
    let user_id = extract_user_id(&req, &jwt_secret).await?;
    
    match repo.get_active(user_id)? {
        Some(entry) => Ok(HttpResponse::Ok().json(entry)),
        None => Ok(HttpResponse::NoContent().finish()),
    }
}

#[delete("/{id}")]
pub async fn delete_time_entry(
    req: actix_web::HttpRequest,
    repo: web::Data<TimeEntryRepository>,
    jwt_secret: web::Data<String>,
    id: web::Path<Uuid>,
) -> Result<HttpResponse, ServiceError> {
    let user_id = extract_user_id(&req, &jwt_secret).await?;
    
    repo.delete(id.into_inner(), user_id)?;
    Ok(HttpResponse::NoContent().finish())
}