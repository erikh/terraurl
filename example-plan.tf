terraform {
  required_providers {
    terraurl = {
      source  = "github.com/erikh/terraurl"
      version = "0.0.1"
    }
  }
}

provider "terraurl" {
  user_agent = "Hello my baby, hello my honey, hello my ragtime gal"
}

resource "terraurl_fetch" "golang" {
  url         = "https://storage.googleapis.com/golang/go1.15.8.linux-amd64.tar.gz"
  target_path = "/tmp/golang.tar.gz"
}
