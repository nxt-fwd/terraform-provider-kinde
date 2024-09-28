resource "kinde_application" "example" {
  name          = "Terraform Acceptance Example Application"
  type          = "reg"
  login_uri     = "http://localhost:3000/oauth/login"
  homepage_uri  = "http://localhost:3000"
  logout_uris   = ["http://localhost:3000/oauth/logout"]
  redirect_uris = ["http://localhost:3000/oauth/redirect"]
}
