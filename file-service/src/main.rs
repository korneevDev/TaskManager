use actix_web::{web, App, HttpServer};
use file_service::{AppConfig, upload_file, download_file, delete_file, list_files, 
                  analyze_text, search_in_text, process_text};
use std::sync::Arc;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let config = Arc::new(AppConfig::from_env());
    
    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::from(config.clone()))
            .service(upload_file)
            .service(download_file)
            .service(delete_file)
            .service(list_files)
            .service(analyze_text)
            .service(search_in_text)
            .service(process_text)
    })
    .bind("127.0.0.1:8082")?
    .run()
    .await
}