
variable "instance_count" {
  default = 1
}

provider "google-beta" {
  credentials = file("/tmp/SSDNIA.json")
  project     = var.project
  region      = var.region
  zone        = var.zone
}

resource "google_compute_instance_from_machine_image" "tpl" {
  provider = google-beta
  project     = var.project
  zone        = var.zone
  count        = var.instance_count

  source_machine_image = "projects/fast-blueprint-296210/global/machineImages/video-converter-image-3-2020-12-30"
  name         = "service-provider-${count.index}"


  metadata = {
    ssh-keys = "${var.gce_ssh_user}:${file(var.gce_ssh_key_location)}"
  }

  connection {
    type        = "ssh"
    user        = var.gce_ssh_user
    private_key = file(var.gce_ssh_private_key_location)
    host = self.network_interface[0].access_config[0].nat_ip
    timeout = "50s"
    agent = false
  }

  provisioner "file" {
    source = "startService.sh"
    destination = "/tmp/startService.sh"
  }

  provisioner "file" {
    source = "/tmp/SSDNIA.json"
    destination = "/tmp/SSDNIA.json"
  }


  provisioner "file" {
    source = "/tmp/MOOGSOFT_KEY.json"
    destination = "/tmp/MOOGSOFT_KEY.json"
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
      "sudo chmod +x /home/group9/CloudVideoConverter/scripts/tfScripts/Service/startService.sh",
      "sudo nohup /home/group9/CloudVideoConverter/scripts/tfScripts/Service/startService.sh &",
      "sleep 1",
    ]
  }

  service_account {
    scopes = ["cloud-platform"]
  }
}