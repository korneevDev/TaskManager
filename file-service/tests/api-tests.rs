use actix_web::{test, web, App, http::StatusCode};
use file_service::{
    text_processor::{TextProcessor, TextProcessingError},
    AppConfig, upload_file, download_file, delete_file, list_files,
    analyze_text, search_in_text, process_text
};
use std::sync::Arc;
use std::fs;
use actix_web::web::Bytes;

#[actix_rt::test]
async fn test_upload_and_download() {
    let app = create_test_app().await;

    let req = test::TestRequest::post()
        .uri("/upload")
        .set_payload(Bytes::from("test content"))
        .to_request();
    let resp = test::call_service(&app, req).await;
    assert_eq!(resp.status(), StatusCode::OK);
    
    let body = test::read_body(resp).await;
    let filename = String::from_utf8(body.to_vec()).unwrap();
    let filename = filename.replace("File saved as: ", "");

    let req = test::TestRequest::get()
        .uri(&format!("/download/{}", filename.split('/').last().unwrap()))
        .to_request();
    let resp = test::call_service(&app, req).await;
    assert_eq!(resp.status(), StatusCode::OK);
}

#[test]
async fn test_text_processor() {
    let content = "test content";
    let temp_file = tempfile::NamedTempFile::new().unwrap();
    fs::write(temp_file.path(), content).unwrap();
    
    let processor = TextProcessor::new(temp_file.path()).unwrap();
    assert_eq!(processor.content(), content);
    assert_eq!(processor.count_words(), 2);
    assert_eq!(processor.count_word_occurrences("test"), 1);
    
    let lines = processor.find_lines_with("test");
    assert_eq!(lines.len(), 1);
    assert_eq!(lines[0], "test content");
}

// Тесты для API endpoints
async fn create_test_app() -> impl actix_web::dev::Service<
    actix_http::Request,
    Response = actix_web::dev::ServiceResponse,
    Error = actix_web::Error
> {
    let config = Arc::new(AppConfig::from_env());
    fs::create_dir_all(&config.upload_dir).unwrap();
    
    test::init_service(
        App::new()
            .app_data(web::Data::from(config))
            .service(upload_file)
            .service(download_file)
            .service(delete_file)
            .service(list_files)
            .service(analyze_text)
            .service(search_in_text)
            .service(process_text)
    ).await
}

#[actix_rt::test]
async fn test_upload_download_flow() {
    let app = create_test_app().await;
    
    // Тестируем загрузку файла
    let req = test::TestRequest::post()
        .uri("/upload")
        .set_payload(Bytes::from("test content"))
        .to_request();
    let resp = test::call_service(&app, req).await;
    assert_eq!(resp.status(), StatusCode::OK);
    
    let body = test::read_body(resp).await;
    let filename = String::from_utf8(body.to_vec()).unwrap();
    let filename = filename.replace("File saved as: ", "");
    let file_id = filename.split('/').last().unwrap();
    
    // Тестируем скачивание файла
    let req = test::TestRequest::get()
        .uri(&format!("/download/{}", file_id))
        .to_request();
    let resp = test::call_service(&app, req).await;
    assert_eq!(resp.status(), StatusCode::OK);
    
    // Тестируем удаление файла
    let req = test::TestRequest::delete()
        .uri(&format!("/delete/{}", file_id))
        .to_request();
    let resp = test::call_service(&app, req).await;
    assert_eq!(resp.status(), StatusCode::OK);
}

#[actix_rt::test]
async fn test_text_analysis() {
    let app = create_test_app().await;
    let config = AppConfig::from_env();
    
    // Создаем тестовый файл в upload_dir
    let test_file = format!("{}/test_analysis.txt", config.upload_dir);
    fs::write(&test_file, "hello world\nhello rust").unwrap();
    
    let req = test::TestRequest::post()
        .uri(&format!("/text/analyze/test_analysis.txt"))
        .to_request();
    let resp = test::call_service(&app, req).await;
    assert_eq!(resp.status(), StatusCode::OK);
    
    let body = test::read_body(resp).await;
    let result: serde_json::Value = serde_json::from_slice(&body).unwrap();
    assert_eq!(result["word_count"], 4);
    
    fs::remove_file(test_file).unwrap();
}

#[actix_rt::test]
async fn test_text_search() {
    let app = create_test_app().await;
    let config = AppConfig::from_env();
    
    // Создаем тестовый файл в upload_dir
    let test_file = format!("{}/test_search.txt", config.upload_dir);
    fs::write(&test_file, "hello world\nhello rust").unwrap();
    
    let req = test::TestRequest::get()
        .uri(&format!("/text/search/test_search.txt?word=hello"))
        .to_request();
    let resp = test::call_service(&app, req).await;
    assert_eq!(resp.status(), StatusCode::OK);
    
    let body = test::read_body(resp).await;
    let result: serde_json::Value = serde_json::from_slice(&body).unwrap();
    assert_eq!(result["count"], 2);
    assert_eq!(result["lines"].as_array().unwrap().len(), 2);
    
    fs::remove_file(test_file).unwrap();
}

#[actix_rt::test]
async fn test_process_text() {
    let app = create_test_app().await;
    
    let req = test::TestRequest::post()
        .uri("/text/process")
        .set_payload("Test")
        .to_request();
    let resp = test::call_service(&app, req).await;
    assert_eq!(resp.status(), StatusCode::OK);
    
    let body = test::read_body(resp).await;
    let result: serde_json::Value = serde_json::from_slice(&body).unwrap();
    assert_eq!(result["upper"], "TEST");
    assert_eq!(result["lower"], "test");
}

#[actix_rt::test]
async fn test_list_files() {
    let app = create_test_app().await;
    
    // Создаем тестовый файл
    let test_file = "./uploads/test_list.txt";
    fs::write(test_file, "test").unwrap();
    
    let req = test::TestRequest::get()
        .uri("/list")
        .to_request();
    let resp = test::call_service(&app, req).await;
    assert_eq!(resp.status(), StatusCode::OK);
    
    let body = test::read_body(resp).await;
    let files: Vec<String> = serde_json::from_slice(&body).unwrap();
    assert!(files.contains(&"test_list.txt".to_string()));
    
    fs::remove_file(test_file).unwrap();
}