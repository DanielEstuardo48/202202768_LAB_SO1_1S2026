use actix_web::{web, App, HttpServer, HttpResponse};
use serde::{Deserialize, Serialize};

#[derive(Deserialize, Serialize, Debug, Clone)]
struct WarReport {
    country: String,
    warplanes_in_air: i32,
    warships_in_water: i32,
    timestamp: String,
}

async fn health() -> HttpResponse {
    HttpResponse::Ok().json(serde_json::json!({"status": "ok"}))
}

async fn receive_report(
    report: web::Json<WarReport>,
    client: web::Data<reqwest::Client>,
) -> HttpResponse {
    println!("Reporte recibido: {:?}", report);

    let go_url = "http://go-client:8081/report";
    
    let result = client
        .post(go_url)
        .json(&report.into_inner())
        .send()
        .await;

    match result {
        Ok(_) => HttpResponse::Ok().json(serde_json::json!({"status": "ok"})),
        Err(e) => {
            eprintln!("Error enviando a Go: {}", e);
            HttpResponse::InternalServerError().json(
                serde_json::json!({"status": "error"})
            )
        }
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let client = reqwest::Client::new();
    
    println!("API Rust corriendo en puerto 8080");

    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(client.clone()))
            .route("/", web::get().to(health))
            .route("/health", web::get().to(health))
            .route("/grpc-202202768", web::post().to(receive_report))
    })
    .bind("0.0.0.0:8080")?
    .run()
    .await
}