use reqwest::blocking::Client;

fn main() {
    let client = Client::new();
    let response = client.get("https://www.example.com").send();
    println!("{:?}", response.expect("REASON").text());    
}
