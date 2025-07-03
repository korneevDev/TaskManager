use actix_web::{test, web, App, http::StatusCode};
use file_service::{upload_file, download_file, delete_file, list_files, 
                  analyze_text, search_in_text, process_text, AppConfig};
use std::sync::Arc;
use std::fs;
use actix_web::web::Bytes;

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