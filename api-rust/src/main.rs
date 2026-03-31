// main.rs: API en Rust que recibe tráfico de Locust y lo reenvía a un Deployment de Go
use actix_web::{post, web, App, HttpResponse, HttpServer, Responder};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use std::env;

// 1. Definimos la estructura JSON que envía Locust
#[derive(Debug, Serialize, Deserialize)]
struct WeatherTweet {
    municipality: i32,
    temperature: i32,
    humidity: i32,
    weather: i32,
}

// 2. Estado de la aplicación para reutilizar el cliente HTTP
struct AppState {
    http_client: Client,
    go_router_url: String,
}

// 3. El endpoint que recibe el tráfico de Locust
#[post("/tweet")]
async fn receive_tweet(
    tweet: web::Json<WeatherTweet>,
    data: web::Data<AppState>,
) -> impl Responder {
    
    // Reenviamos el JSON exactamente como llegó al Deployment de Go
    let res = data.http_client
        .post(&data.go_router_url)
        .json(&tweet.into_inner())
        .send()
        .await;

    match res {
        Ok(response) if response.status().is_success() => {
            HttpResponse::Ok().body("Tweet procesado correctamente")
        }
        Ok(_) => HttpResponse::BadGateway().body("El servidor Go respondió con error"),
        Err(_) => HttpResponse::InternalServerError().body("Error de conexión con Go"),
    }
}

// 4. Configuración principal del servidor
#[actix_web::main]
async fn main() -> std::io::Result<()> {
    // Definimos la URL de Go (por defecto localhost:8080 para pruebas locales)
    let go_router_url = env::var("GO_ROUTER_URL")
        .unwrap_or_else(|_| "http://localhost:8080/tweet".to_string());

    // Instanciamos el cliente HTTP una sola vez para mayor rendimiento
    let app_state = web::Data::new(AppState {
        http_client: Client::new(),
        go_router_url,
    });

    println!("Iniciando API Rust en el puerto 8000...");

    HttpServer::new(move || {
        App::new()
            .app_data(app_state.clone())
            .service(receive_tweet)
    })
    .bind(("0.0.0.0", 8000))?
    .workers(4) // Optimizamos los hilos para soportar la alta carga de Locust
    .run()
    .await
}