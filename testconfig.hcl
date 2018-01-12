adapter "deconz" {
    type = "deconz"
    config {
        hostname = "10.0.1.22"
        port = 80
    }
}

adapter "dra" {
    type = "dra"
    config {
        address = "10.0.1.8:23"
    }
}

trigger "toggle_all_light_on" {
    key = "deconz.sensors.10.buttonevent"
    value = 5002
    target = "deconz.groups.1.on"
    target_value = false
}

trigger "toggle_denon_on" {
    key = "deconz.sensors.10.buttonevent"
    value = 4002
    target = "dra.on"
    target_value = false
}
