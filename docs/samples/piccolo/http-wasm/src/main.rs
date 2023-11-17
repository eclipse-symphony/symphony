use std::net::SocketAddr;
use hyper::server::conn::Http;
use hyper::service::service_fn;
use hyper::{Body, Method, Request, Response};
use tokio::net::TcpListener;
use std::fs;


async fn echo(req: Request<Body>) -> Result<Response<Body>, hyper::Error> {
  let contents = match fs::read_to_string("/trace_pipe") {
    Ok(contents) => format!("<h1>Web server inside WebAssembly</h1><p>eBPF traces:<p>{}</p></p>", contents),
    Err(_) => format!("<h1>Web server inside WebAssembly</h1><p>No traces</p></p>"),
  };

  match (req.method(), req.uri().path()) {
    
    (&Method::GET, "/") => Ok(Response::new(Body::from(contents))),    

    _ => {
        Ok(Response::new(Body::from("Not found")))
    }
  }
}

#[tokio::main(flavor = "current_thread")]
async fn main() -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
  let addr = SocketAddr::from(([0, 0, 0, 0], 8085));

  let listener = TcpListener::bind(addr).await?;
  println!("Listening on http://{}", addr);
  loop {
    let (stream, _) = listener.accept().await?;

    tokio::task::spawn(async move {
        if let Err(err) = Http::new().serve_connection(stream, service_fn(echo)).await {
          println!("Error serving connection: {:?}", err);
        }
    });
  }
}