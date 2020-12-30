
variable "instance_count" {
  default = 1
}

provider "google-beta" {
  credentials = file("SSDNIA.json")
  project     = var.project
  region      = var.region
  zone        = var.zone
}

resource "google_compute_instance_from_machine_image" "tpl" {
  provider = google-beta
  project     = var.project
  zone        = var.zone
  count        = var.instance_count
  name         = "load-balancer-${count.index}"
  source_machine_image = "projects/fast-blueprint-296210/global/machineImages/video-converter-image-2-0"


  metadata = {
    ssh-keys = "${var.gce_ssh_user}:${file(var.gce_ssh_key_location)}"
  }

  connection {
    type        = "ssh"
    user        = var.gce_ssh_user
    private_key = file(var.gce_ssh_private_key_location)
    host = self.network_interface[0].access_config[0].nat_ip
    timeout = "20s"
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
      "sudo chmod +x /tmp/*",
      "sudo /tmp/startAPIGateway.sh",
      "sudo nohup /tmp/startLoadBalancer.sh &",
      "sleep 1"
    ]
  }
  service_account {
    scopes = ["cloud-platform"]
  }
}