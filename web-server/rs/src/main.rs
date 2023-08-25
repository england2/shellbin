use actix_session::{storage::SessionStore, Session, SessionMiddleware};
use awc::Client;
use url::Url;

// use actix_web::{get, post, web, App, Error, HttpRequest, HttpResponse, HttpServer, Responder};
use actix_files::{Files, NamedFile};
use actix_web::{
    error, get,
    http::{
        header::{self, ContentType},
        Method, StatusCode,
    },
    middleware, post, web, App, Either, HttpRequest, HttpResponse, HttpServer, Responder, Result,
};

#[get("/")]
async fn hello(req: HttpRequest, session: Session) -> impl Responder {
    println!("{req:?}");
    let uri = req.path();

    println!("{uri:?}");

    // TODO frontend landing page
    HttpResponse::Ok().body("Shellbin")
}

// this function to be called asynchronously (?)
async fn get_paste(req: HttpRequest, session: Session) -> Result<HttpResponse> {
    let path = req.path();
   
    println!("{path:?}"); //tk

    
    //TODO? refactor http get to reusable function
    // is this async?? I want this to block
    let resp = match reqwest::get("https://httpbin.org/ip").await {
        Ok(resp) => resp.text().await.unwrap(),
        Err(err) => panic!("Error: {}", err)
    };
    println!("{}", resp);
    

    let body = format!("{}", path);

    Ok(HttpResponse::build(StatusCode::OK)
        .content_type(ContentType::plaintext())
        .body(body))
}

#[post("/echo")]
async fn echo(req_body: String) -> impl Responder {
    HttpResponse::Ok().body(format!("from server: {req_body}"))
}

async fn manual_hello() -> impl Responder {
    HttpResponse::Ok().body("Hey there!")
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let addr = "127.0.0.1:4561";

    println!("running on: {:?}", addr);

    HttpServer::new(|| {
        App::new()
            .service(hello)
            .service(echo)
            .route("/hey", web::get().to(manual_hello))
            // .default_service(web::to(default_handler))
            .default_service(web::to(get_paste))
    })
    .bind(addr)?
    // .bind(("127.0.0.1", 4560))?
    .run()
    .await
}
