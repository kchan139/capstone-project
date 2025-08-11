resource "digitalocean_firewall" "capstone" {
  name = "capstone-firewall"

  droplet_ids = [digitalocean_droplet.capstone.id]

  ### SSH Access ###

  # Allow SSH on default port from trusted IPs.
  # This rule is meant for initial Ansible provisioning.
  # You should lock this port down afterward 
  # 
  inbound_rule {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = ["0.0.0.0/0", "::/0"]
    # source_addresses = var.ssh_access_ips
  }

  inbound_rule {
    protocol         = "tcp"
    port_range       = var.custom_ssh_port
    source_addresses = ["0.0.0.0/0", "::/0"]
    # source_addresses = var.ssh_access_ips
  }

  ### ICMP for Ping ###
  inbound_rule {
    protocol         = "icmp"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  ### Outbound Rules (Left open for now) ###
  outbound_rule {
    protocol              = "tcp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "udp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "icmp"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}
