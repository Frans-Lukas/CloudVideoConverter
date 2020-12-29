
variable "instance_count" {
  default = 1
}

//command = "echo $(pwd)"

provider "google" {
  credentials = file("../LoadBalancer/SSDNIA.json")
  project     = var.project
  region      = var.region
  zone        = var.zone
}

resource "google_compute_instance" "vm_instance" {
  count        = var.instance_count
  name         = "spawning-pool-${count.index}"
  machine_type = "f1-micro"
  tags = ["http-server", "https-server"]

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
    source = "startSpawningPool.sh"
    destination = "/tmp/startSpawningPool.sh"
  }

  provisioner "file" {
    source = "../LoadBalancer/SSDNIA.json"
    destination = "/tmp/SSDNIA.json"
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
      "nohup /tmp/startSpawningPool.sh &",
      "sleep 1",
    ]
  }
}