output "capstone_droplet_ip" {
  description = "DigitalOcean droplet IP for the Capstone Project"
  value       = digitalocean_droplet.capstone.ipv4_address
}
