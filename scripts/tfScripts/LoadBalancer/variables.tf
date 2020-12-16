variable "project" {
  description = "Project ID"
  default     = "fast-blueprint-296210"
}

variable "region" {
  description = "Region name"
  default     = "europe-north1"
}

variable "zone" {
  description = "Zone name"
  default     = "europe-north1-a"
}

variable "gce_ssh_user" {
  default = "group9"
}

variable "gce_ssh_key_location" {
  default = "/tmp/id_rsa.pub"
}

variable "gce_ssh_private_key_location" {
  default = "/tmp/id_rsa"
}