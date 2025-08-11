variable "khoa_ssh_public_key" {
  description = "Khoa's SSH Public Key"
  type        = string
}

variable "phiung_ssh_public_key" {
  description = "Phi Ung's SSH Public Key"
  type        = string
}

variable "custom_ssh_port" {
  description = "Custom SSH Port"
  type        = string
}

variable "ssh_access_ips" {
  description = "List of CIDR blocks allowed to access SSH"
  type        = list(string)
  sensitive   = true
}

variable "do_token" {
  description = "DigitalOcean API Token"
  type        = string
  sensitive   = true
}
