resource "digitalocean_ssh_key" "khoa" {
  name       = "Khoa's SSH key"
  public_key = var.khoa_ssh_public_key
}

resource "digitalocean_ssh_key" "phiung" {
  name       = "Phi Ung's SSH key"
  public_key = var.phiung_ssh_public_key
}

resource "digitalocean_droplet" "capstone" {
  image  = "ubuntu-24-04-x64"
  name   = "capstone-project"
  region = "sgp1"
  size   = "s-2vcpu-4gb"
  ssh_keys = [
    digitalocean_ssh_key.khoa.id,
    digitalocean_ssh_key.phiung.id,
  ]

  # provisioner "local-exec" {
  #   command     = "sleep 100 && ./run.sh"
  #   working_dir = "../ansible/scripts"
  #   when        = create
  # }

  backups = false
  # backup_policy {
  #   plan    = "weekly"
  #   weekday = "TUE"
  #   hour    = 8
  # }
}
