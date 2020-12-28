
variable "instance_count" {
  default = 1
}

provider "google" {
  credentials = file("SSDNIA.json")
  project     = var.project
  region      = var.region
  zone        = var.zone
}

resource "google_compute_instance" "vm_instance" {
  count        = var.instance_count
  name         = "load-balancer-${count.index}"
  machine_type = "f1-micro"


  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-9"
    }
  }

  network_interface {
    network = "default"
    access_config {
    }
  }

  metadata = {
    ssh-keys = "${var.gce_ssh_user}:${file(var.gce_ssh_key_location)}"
  }

  connection {
    type        = "ssh"
    user        = var.gce_ssh_user
    private_key = file(var.gce_ssh_private_key_location)
    host = self.network_interface[0].access_config[0].nat_ip
    timeout = "10s"
    agent = false
  }

  provisioner "file" {
    source = "startLoadBalancer.sh"
    destination = "/tmp/startLoadBalancer.sh"
  }

  provisioner "file" {
    source = "SSDNIA.json"
    destination = "/tmp/SSDNIA.json"
  }

  provisioner "file" {
    source = "startEverythingElse.sh"
    destination = "/tmp/startEverythingElse.sh"
  }
  provisioner "file" {
    source = "startAPIGateway.sh"
    destination = "/tmp/startAPIGateway.sh"
  }

  provisioner "file" {
    source = "/tmp/id_rsa.pub"
    destination = "/tmp/id_rsa.pub"
  }

  provisioner "file" {
    source = "/tmp/id_rsa"
    destination = "/tmp/id_rsa"
  }

  provisioner "remote-exec" {
    inline = [
      "chmod +x /tmp/*",
      "/tmp/startAPIGateway.sh",
      "nohup /tmp/startLoadBalancer.sh 50051 &",
      "/tmp/startEverythingElse.sh",
    ]
  }
}