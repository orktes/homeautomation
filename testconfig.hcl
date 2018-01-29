# Templates

{{define "alexa_deconz_lightgroup"}}
device "{{slugify (index . 0)}}" {
    name = "{{index . 0}}"
    description = "{{index . 0}} group"

    display_categories = ["SWITCH"]

    capability "PowerController" {
        property "powerState" {
            get = "get('haaga/deconz/groups/{{index . 1}}/any_on') ? 'ON' : 'OFF'"
            set = "set('haaga/deconz/groups/{{index . 1}}/on', value === 'ON')"
        }
    }
}
{{end}}


# Actual config 

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

    {{template "alexa_deconz_lightgroup" array "All lights" 3 }}
    {{template "alexa_deconz_lightgroup" array "Bedroom lights" 4 }}
    {{template "alexa_deconz_lightgroup" array "Kitchen lights" 2 }}
    {{template "alexa_deconz_lightgroup" array "Living room lights" 1 }}

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