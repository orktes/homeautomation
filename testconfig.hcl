servers = ["tcp://localhost:1883"]

bridge {
    root = "haaga"

    adapter "deconz" {
        type = "deconz"
        config {
            hostname = "10.0.1.22"
            port = 80
            key = "624F099FA9"
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
    

    adapter "db" {
        type = "bolt"
        config {
            database_file = "./test.db"
        }
    } 
}


trigger {
    script = <<SOURCE
        listen("haaga/dra/master_volume", function () {
            print("dra volume", get("haaga/dra/master_volume"), "power", get("haaga/dra/power"))
        });

        subscribe("haaga/#", function (topic) {
            print("Wildcard received", topic)
        })
    SOURCE
}