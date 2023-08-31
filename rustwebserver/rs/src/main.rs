use actix_files::{Files, NamedFile};
use actix_session::{storage::SessionStore, Session, SessionMiddleware};
use actix_web::{
    error, get,
    http::{
        header::{self, ContentType},
        Method, StatusCode,
    },
    middleware, post,
    web::{self, post},
    App, Either, HttpRequest, HttpResponse, HttpServer, Responder, ResponseError, Result,
};
use awc::Client;
use derive_more::{Display, Error};
use reqwest::Response;
use serde::{Deserialize, Serialize};
use std::env;
use url::Url;
#[macro_use]
extern crate lazy_static;

lazy_static! {
    static ref CLIENT: reqwest::Client = reqwest::Client::new();
    static ref DBSERVICEADDR: String =
        env::var("DBSERVICEADDR").expect("Error: DBSERVICEADDR not found");
    static ref HOSTADDR: String = env::var("HOSTADDR").expect("Error: HOSTADDR not found");
}

#[allow(non_snake_case)]
#[derive(Deserialize, Debug)]
struct Paste {
    Hash: String,
    Content: String,
    Created: String,
    LastAccessed: String,
}

// TODO frontend
async fn index(req: HttpRequest, session: Session) -> impl Responder {
    HttpResponse::Ok().body("Shellbin")
}

async fn default_route(req: HttpRequest) -> Result<HttpResponse> {
    let path = req.path();

    let post_result = CLIENT
        .post("http://".to_string() + &DBSERVICEADDR.to_string() + "/servePaste")
        .json(&serde_json::json!({
           "Hash": path,
        }))
        .send()
        .await;

    // error if db-service unreachable
    if post_result.is_err() {
        println!("=== ERROR CALLING DB SERVICE ===");
        println!("{:?}", post_result.unwrap_err());

        // TODO learn actix web error. return Err(something)
        return Ok(HttpResponse::build(StatusCode::INTERNAL_SERVER_ERROR)
            .content_type(ContentType::plaintext())
            .body("INTERNAL SERVER ERROR DB-SERVICE")); //t
    }

    let response = post_result.unwrap();
    println!("=== RESPONSE ==="); //t
    println!("{:?}", response); //t

    let binding = response.status();
    let code = binding.as_str();
    match code {
        "200" => {
            println!("=== SUCCESS 200 ===");
            match response.json::<Paste>().await {
                Ok(paste) => {
                    println!("{:?}", paste);
                    return Ok(HttpResponse::build(StatusCode::OK)
                        .content_type(ContentType::plaintext())
                        .body(paste.Content)); //t
                }
                Err(err) => {
                    println!("=== ERROR DESERIALIZING TO JSON ===");
                    println!("{:?}", err);
                    // TODO return error
                    return Ok(HttpResponse::build(StatusCode::OK)
                        .content_type(ContentType::plaintext())
                        .body("INTERNAL SERVER ERROR JSON")); //t
                }
            }
        }
        "404" => {
            println!("{}: Error not found", code);
            return Ok(HttpResponse::Ok().body("error 404: paste not found"));
        }
        &_ => {
            println!("{}: Code unknown", code);
            return Ok(HttpResponse::Ok().body("error { code }"));
        }
    }
}

#[post("/echo")]
async fn echo(req_body: String) -> impl Responder {
    HttpResponse::Ok().body(format!("from server: {req_body}"))
}

async fn manual_hello() -> impl Responder {
    HttpResponse::Ok().body("Hey there!")
}

//t
async fn default() -> impl Responder {
    println!("in default");
    HttpResponse::Ok().body("default")
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    println!("{}", DBSERVICEADDR.to_string()); //t
    let addr: String = HOSTADDR.to_string();
    println!("running on: {:?}", addr);

    HttpServer::new(|| {
        App::new()
            .service(echo)
            .route("/", web::get().to(index))
            .route("/hey", web::get().to(manual_hello))
            .default_service(web::to(default_route))
        // .default_service(web::to(default)) //t
    })
    .bind(addr)?
    .run()
    .await
}
