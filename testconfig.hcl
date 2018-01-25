root = "haaga"
servers = ["tcp://localhost:1883"]

/*
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
*/

adapter "db" {
    type = "bolt"
    config {
        database_file = "./test.db"
    }
} 