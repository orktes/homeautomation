frontend "ssh" {
    type = "ssh"
    config {
        addr = ":2222"
        password = "jaakonkoti"
    }
}

adapter "deconz" {
    type = "deconz"
    config {
        hostname = "10.0.1.22"
        port = 80
        key = "5D7E8C715E"
    }
}

adapter "dra" {
    type = "dra"
    config {
        address = "10.0.1.8:23"
    }
}

adapter "tv" {
    type = "viera"
    config {
        mac = "48:A9:D2:53:DC:10"
    }
}

trigger "denon_volume_up" {
    key = "deconz.sensors[10].buttonevent"
    condition = "dra.power && deconz.sensors[10].buttonevent === 5001"
    end_condition = "deconz.sensors[10].buttonevent === 5003"
    action = "dra.master_volume = 'UP'"
    interval = 100
}

trigger "denon_volume_down" {
    key = "deconz.sensors[10].buttonevent"
    condition = "dra.power && deconz.sensors[10].buttonevent === 4001"
    end_condition = "deconz.sensors[10].buttonevent === 4003"
    action = "dra.master_volume = 'DOWN'"
    interval = 100
}

light "Livingroom" {
    read {
        bri = "deconz.groups.1.bri"
        on = "deconz.groups.1.any_on"
    }
    write {
        bri = "deconz.groups.1.bri"
        on = "deconz.groups.1.on"
    }
}