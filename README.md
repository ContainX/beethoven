# B E E T H O V E N

Is a Mesos/Marathon stream listener which manages an NGINX proxy instance within Docker.  It allows you to extend a the docker image and add an `nginx.conf.tmpl` which will be dynamically parsed into the final configuration file based on state changes within the cluster.

This is a similar solution to `marathon-lb` but for the HTTP layer offering NGINX's advance http level uri routing.

## Overview

![Architecture](images/architecture.jpg?raw=true "Architecture")

Beethoven runs in your cluster as a container managed by Marathon.  This allows for horizontal scaling across the Mesos agents.  Supervisor is leveraged to manage both Nginx and Beethoven since we are running two executables in a single container.

Beethoven attaches to the SSE stream in Marathon and listens for real-time changes to Apps and Tasks.  This includes new applications being added to the cluster or existing which has had a state change such as a health event.

When an event occurs Beethoven takes the user provided `nginx.template` and parses it with the Handlebars processor.  Handlebars offers a lot of power behind the template including logic blocks, object iteration and other conditional behaviours. 

After the template has been parsed a temp Nginx configuration file is rendered within the container.  Beethoven then asks Nginx to validate the temporary configuration file to determine if it's syntax is correct.  If the syntax is good then the current `nginx.conf` is replaced with the temp. file and a soft reload is issued.  If the temp file is bad then it is recorded and associated with the `/bt/status/` endpoint for debugging.

**Feature Highlights**

* Uses Nginx for HTTP based loadbalancing
* Handlebars for powerful template parsing
* Listens to the realtime SSE from Marathon to quickly change upstreams based on application/tasks state changes
* RESTful endpoints for current status
* Flexible configuration options (local config, spring-cloud configuration remote configuration fetching and ENV variables)
* Easy to get started add a `FROM containx/beethoven` to your `Dockerfile` add your template, config options and release!

## Getting Started

TODO DOC

## License

This software is licensed under the Apache 2 license, quoted below.

Copyright 2016 ContainX / Jeremy Unruh

Licensed under the Apache License, Version 2.0 (the "License"); you may not
use this file except in compliance with the License. You may obtain a copy of
the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
License for the specific language governing permissions and limitations under
the License.