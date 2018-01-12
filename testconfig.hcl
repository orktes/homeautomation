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

trigger "test" {
    key = "deconz.sensors.10.buttonevent"
    value = 5002
    target = "deconz.groups.1.on"
    target_value = false
}

trigger "test2" {
    key = "deconz.sensors.10.buttonevent"
    value = 4002
    target = "deconz.lights.7.on"
    target_value = false
}
