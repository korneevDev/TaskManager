use actix_web::{web, App, HttpServer};
use actix_cors::Cors;
use dotenv::dotenv;
use std::env;

mod models;
mod schema;
mod handlers;
mod repository;
mod errors;

use repository::TimeEntryRepository;
use handlers::{time_entry_handlers, health_check};

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    dotenv().ok();
    env_logger::init();

    let database_url = env::var("DATABASE_URL").expect("DATABASE_URL must be set");
    let jwt_secret = env::var("JWT_SECRET").expect("JWT_SECRET must be set");
    let repo = TimeEntryRepository::new(&database_url);
    
    let host = env::var("HOST").unwrap_or_else(|_| "127.0.0.1".to_string());
    let port = env::var("PORT").unwrap_or_else(|_| "8083".to_string());
    let bind_address = format!("{}:{}", host, port);

    println!("Starting server at {}", bind_address);

    HttpServer::new(move || {
        let cors = Cors::default()
            .allow_any_origin()
            .allowed_methods(vec!["GET", "POST", "PUT", "DELETE"])
            .allowed_headers(vec!["Authorization", "Content-Type"])
            .max_age(3600);

        App::new()
            .wrap(cors)
            .app_data(web::Data::new(repo.clone()))
            .app_data(web::Data::new(jwt_secret.clone()))
            .service(health_check)
            .configure(time_entry_handlers)
    })
    .bind(&bind_address)?
    .run()
    .await
}