
variable "instance_count" {
  default = 1
}

//command = "echo $(pwd)"
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
  source_machine_image = "projects/fast-blueprint-296210/global/machineImages/video-converter-image-1-2020-12-29"
  name         = "spawning-pool-${count.index}"

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