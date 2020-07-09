# Factom Eater

A factomd live feed API consumer library intended to provide a lightweight endpoint. This currently does not handle errors well and is not great for debugging purposes. 

## Usage

Create a new Eater with `eater.Launch(host)`, where `host` is the endpoint specified in *factomd.conf*. If the factomd setting is the default, you can use `Launch(":8040")`. Launching will return an error if the app is unable to listen at the specified address. The launcher can be stopped via `Stop()`, which will sever all active connections and stop listening for new ones. 
