# MQTT Brokers
servers = ["tcp://localhost:1883"]

bridge {
    # MQTT topic root path
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
    {{ define "speaker_capability" }}
    capability "Speaker" {
        property "mute" {
            get = "get('{{index . 0}}')"
            set = "set('{{index . 0}}', value)"
        }

        property "volume" {
            type = "int"
            input_range = [0, 100]
            output_range = [{{index . 2}}, {{index . 3}}]

            get = "get('{{index . 1}}')"
            set = "set('{{index . 1}}', value)"
        } 
    }

    capability "PercentageController" {
        property "percentage" {
            type = "int"
            input_range = [0, 100]
            output_range = [{{index . 2}}, {{index . 3}}]

            get = "get('{{index . 1}}')"
            set = "set('{{index . 1}}', value)"
        } 
    }
    {{end}}

    {{ define "temperature_ranges" }}
        type = "int"
        input_range = [1000, 10000]
        output_range = [152, 353]
    {{end}}
    {{define "alexa_deconz_lightgroup"}}
    device "{{slugify (index . 0)}}" {
        name = "{{index . 0}}"
        description = "{{index . 0}} group"

        display_categories = ["LIGHT"]

        capability "PowerController" {
            property "powerState" {
                get = "get('haaga/deconz/groups/{{index . 1}}/any_on') ? 'ON' : 'OFF'"
                set = "set('haaga/deconz/groups/{{index . 1}}/on', value === 'ON')"
            }
        }

        capability "BrightnessController" {
            property "brightness" {
                type = "int"
                input_range = [0, 100]
                output_range = [0, 255]

                get = "get('haaga/deconz/groups/{{index . 1}}/bri')"
                set = "set('haaga/deconz/groups/{{index . 1}}/bri', value)"
            }
        }

        capability "ColorTemperatureController" {
            {{ $temperatureStep := 100 }}
            

            property "colorTemperatureInKelvin" {
                {{template "temperature_ranges"}}

                get = "get('haaga/deconz/groups/{{index . 1}}/ct')"
                set = "set('haaga/deconz/groups/{{index . 1}}/ct', value)"
            }

            action "DecreaseColorTemperature" {
                {{template "temperature_ranges"}}

                script = <<SOURCE
                    (function () {
                        var ct = get('haaga/deconz/groups/{{index . 1}}/ct');
                        ct += {{$temperatureStep}};
                        ct = Math.min(ct, 353)
                        set('haaga/deconz/groups/{{index . 1}}/ct', ct);
                        return ct;
                    })();
                SOURCE
            }

            action "IncreaseColorTemperature" {
                {{template "temperature_ranges"}}

                script = <<SOURCE
                    (function () {
                        var ct = get('haaga/deconz/groups/{{index . 1}}/ct');
                        ct -= {{$temperatureStep}};
                        ct = Math.max(ct, 152)
                        set('haaga/deconz/groups/{{index . 1}}/ct', ct);
                        return ct;
                    })();
                SOURCE
            }
        }
    }
    {{end}}

    {{template "alexa_deconz_lightgroup" array "All lights" 3 }}
    {{template "alexa_deconz_lightgroup" array "Bedroom lights" 4 }}
    {{template "alexa_deconz_lightgroup" array "Kitchen lights" 2 }}
    {{template "alexa_deconz_lightgroup" array "Living room lights" 1 }}

    device "livingroom_amp" {
        name = "Amplifier"
        description = "Living room amplifier"

        display_categories = ["SPEAKER"]

        {{template "speaker_capability" array "haaga/dra/mute" "haaga/dra/master_volume" 90 0 }}


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

        display_categories = ["TV"]

        {{template "speaker_capability" array "haaga/tv/1/mute" "haaga/tv/1/volume" 0 100 }}

        capability "PowerController" {
            property "powerState" {
                get = "get('haaga/tv/1/power') ? 'ON' : 'OFF'"
                set = "set('haaga/tv/1/power', value === 'ON')"
            }
        }
    }
}