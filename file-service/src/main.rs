pub mod text_processor;

use actix_web::{
    delete, get, post,
    web::{self, Bytes, Data},
    App, HttpResponse, HttpServer, Responder,
};
use serde_json::json;
use std::{
    collections::HashMap,
    fs::{self, File},
    io::Write,
    path::{Path, PathBuf},
    sync::Arc,
};
use text_processor::{TextProcessingError, TextProcessor};
use uuid::Uuid;

#[derive(Clone)]
struct AppConfig {
    upload_dir: String,
    max_file_size: usize,
}

impl AppConfig {
    fn from_env() -> Self {
        dotenv::dotenv().ok();
        
        AppConfig {
            upload_dir: std::env::var("UPLOAD_DIR").unwrap_or_else(|_| "./uploads".to_string()),
            max_file_size: std::env::var("MAX_FILE_SIZE")
                .unwrap_or_else(|_| "10485760".to_string()) // 10MB по умолчанию
                .parse()
                .expect("MAX_FILE_SIZE must be a number"),
        }
    }
}

const UPLOAD_DIR: &str = "./uploads";

#[post("/upload")]
async fn upload_file(file: Bytes) -> impl Responder {
    if !Path::new(UPLOAD_DIR).exists() {
        fs::create_dir(UPLOAD_DIR).expect("Failed to create upload directory");
    }

    let filename = format!("{}/{}.dat", UPLOAD_DIR, Uuid::new_v4());
    
    match File::create(&filename) {
        Ok(mut f) => {
            if let Err(e) = f.write_all(&file) {
                return HttpResponse::InternalServerError().body(format!("Failed to write file: {}", e));
            }
            HttpResponse::Ok().body(format!("File saved as: {}", filename))
        }
        Err(e) => HttpResponse::InternalServerError().body(format!("Failed to create file: {}", e)),
    }
}

#[post("/text/analyze/{filename}")]
async fn analyze_text(
    filename: web::Path<String>,
    config: Data<AppConfig>,
) -> impl Responder {
    let path = PathBuf::from(&config.upload_dir).join(&*filename);
    
    match TextProcessor::new(&path) {
        Ok(processor) => {
            let word_count = processor.count_words();
            let content_preview = if processor.content().len() > 100 {
                format!("{}...", &processor.content()[..100])
            } else {
                processor.content().to_string()
            };
            
            HttpResponse::Ok().json(json!({
                "filename": &*filename,
                "word_count": word_count,
                "preview": content_preview,
            }))
        }
        Err(e) => match e {
            TextProcessingError::FileNotFound => HttpResponse::NotFound().body("File not found"),
            _ => HttpResponse::InternalServerError().body("Error processing file"),
        },
    }
}

#[get("/text/search/{filename}")]
async fn search_in_text(
    filename: web::Path<String>,
    query: web::Query<HashMap<String, String>>,
    config: Data<AppConfig>,
) -> impl Responder {
    let path = PathBuf::from(&config.upload_dir).join(&*filename);
    
    if let Some(word) = query.get("word") {
        match TextProcessor::new(&path) {
            Ok(processor) => {
                let count = processor.count_word_occurrences(word);
                let lines = processor.find_lines_with(word);
                
                HttpResponse::Ok().json(json!({
                    "filename": &*filename,
                    "word": word,
                    "count": count,
                    "lines": lines,
                }))
            }
            Err(e) => match e {
                TextProcessingError::FileNotFound => HttpResponse::NotFound().body("File not found"),
                _ => HttpResponse::InternalServerError().body("Error processing file"),
            },
        }
    } else {
        HttpResponse::BadRequest().body("Missing 'word' query parameter")
    }
}

#[post("/text/process")]
async fn process_text(
    text: String, 
) -> impl Responder {
    let upper = text_processor::process_text_owned(text);
    let lower = text_processor::process_text_borrowed(&upper);
    
    HttpResponse::Ok().json(json!({
        "original_length": upper.len(),
        "upper": upper,
        "lower": lower,
    }))
}

#[get("/download/{filename}")]
async fn download_file(filename: web::Path<String>) -> impl Responder {
    let path = format!("{}/{}", UPLOAD_DIR, filename);
    
    match fs::read(&path) {
        Ok(content) => HttpResponse::Ok()
            .content_type("application/octet-stream")
            .body(content),
        Err(_) => HttpResponse::NotFound().body("File not found"),
    }
}

#[delete("/delete/{filename}")]
async fn delete_file(filename: web::Path<String>) -> impl Responder {
    let path = format!("{}/{}", UPLOAD_DIR, filename);
    
    match fs::remove_file(&path) {
        Ok(_) => HttpResponse::Ok().body("File deleted"),
        Err(_) => HttpResponse::NotFound().body("File not found"),
    }
}

#[get("/list")]
async fn list_files() -> impl Responder {
    match fs::read_dir(UPLOAD_DIR) {
        Ok(entries) => {
            let files: Vec<String> = entries
                .filter_map(|e| e.ok())
                .map(|e| e.file_name().to_string_lossy().into_owned())
                .collect();
            HttpResponse::Ok().json(files)
        }
        Err(_) => HttpResponse::InternalServerError().body("Failed to read directory"),
    }
}

fn main() -> std::io::Result<()> {
    std::env::set_var("RUST_LOG", "actix_web=info");
    env_logger::init();
    
    let config = Arc::new(AppConfig::from_env());
    
    if !Path::new(&config.upload_dir).exists() {
        fs::create_dir_all(&config.upload_dir)?;
    }

    println!("Server started at http://localhost:8082");
    println!("Upload directory: {}", config.upload_dir);
    println!("Max file size: {} bytes", config.max_file_size);
    
    let rt = tokio::runtime::Runtime::new()?;
    rt.block_on(async {
        HttpServer::new(move || {
            App::new()
                .app_data(Data::from(config.clone()))
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
    })
}