use std::net::SocketAddr;
use hyper::server::conn::Http;
use hyper::service::service_fn;
use hyper::{Body, Method, Request, Response};
use tokio::net::TcpListener;
use std::fs::File;
use std::io::Read;
use std::io::BufReader;

async fn echo(req: Request<Body>) -> Result<Response<Body>, hyper::Error> {
  
  let response = "<h1>Web server inside WebAssembly</h1><a href='/trace'>eBFP Traces</a>";


  match (req.method(), req.uri().path()) {
    
    (&Method::GET, "/") => Ok(Response::new(Body::from(response))),    
    (&Method::GET, "/trace") => {
      let mut contents = String::new();
      let file = match File::open("/trace_pipe") {
        Ok(file) => file,
        Err(_) => return Ok(Response::new(Body::from("<h1>Web server inside WebAssembly</h1><p>No traces</p></p>"))),
      };
      let mut reader = BufReader::new(file).take(1024);
      let _ = reader.read_to_string(&mut contents);
      let resonse = format!("<h1>Web server inside WebAssembly</h1><p>eBFP Traces:</p><pre>{}</pre>", contents);
      Ok(Response::new(Body::from(resonse)))
    },
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