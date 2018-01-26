bridge {
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
}


trigger {
    source = <<SOURCE
        // Listen to smarthome
        listen("haaga/tv/1/power", function (topic, payload) {
            set("haaga/dra/power", true);
            var val = get("haaga/tv/1/volume");
            set("haaga/dra/master_volume", val);
            set("haaga/tv/1/volume", 0);
        });

        // Subscribe to MQTT topic (with QoS 0)
        subscribe("haaga/connected", 0, function (topic, payload) {
            // Payload is just a string here
            publish("haaga/is/online", 0, "Some payload");
        });
    SOURCE
}