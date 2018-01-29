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
/*
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
*/

alexa {
    topic = "haaga/aws/lambda/homeautomation"

    device "all_light" {
        name = "All lights"
        description = "All lights in the appartment"

        display_categories = ["SWITCH"]

        capability "PowerController" {
            property "powerState" {
                get = "get('haaga/deconz/groups/3/any_on') ? 'ON' : 'OFF'"
                set = "set('haaga/deconz/groups/3/on', value === 'ON')"
            }
        }
    }

    device "bedroom_light" {
        name = "Bedroom lights"
        description = "Bedroom lights"

        display_categories = ["SWITCH"]

        capability "PowerController" {
            property "powerState" {
                get = "get('haaga/deconz/groups/4/any_on') ? 'ON' : 'OFF'"
                set = "set('haaga/deconz/groups/4/on', value === 'ON')"
            }
        }
    }

    device "kitchen_light" {
        name = "Kitchen lights"
        description = "Kitchen lights"

        display_categories = ["SWITCH"]

        capability "PowerController" {
            property "powerState" {
                get = "get('haaga/deconz/groups/2/any_on') ? 'ON' : 'OFF'"
                set = "set('haaga/deconz/groups/2/on', value === 'ON')"
            }
        }
    }

    device "livingroom_light" {
        name = "Living room lights"
        description = "Living room lights"

        display_categories = ["SWITCH"]

        capability "PowerController" {
            property "powerState" {
                get = "get('haaga/deconz/groups/1/any_on') ? 'ON' : 'OFF'"
                set = "set('haaga/deconz/groups/1/on', value === 'ON')"
            }
        }
    }

    device "livingroom_amp" {
        name = "Amplifier"
        description = "Living room amplifier"

        display_categories = ["SWITCH"]

        capability "PowerController" {
            property "powerState" {
                get = "get('haaga/dra/power') ? 'ON' : 'OFF'"
                set = "set('haaga/dra/power', value === 'ON')"
            }
        }
    }

    device "livingroom_tv" {
        name = "TV"
        description = "Living room TV"

        display_categories = ["SWITCH"]

        capability "PowerController" {
            property "powerState" {
                get = "get('haaga/tv/1/power') ? 'ON' : 'OFF'"
                set = "set('haaga/tv/1/power', value === 'ON')"
            }
        }
    }
}