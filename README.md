# B E E T H O V E N

Is a Mesos/Marathon stream listener which manages an NGINX proxy instance within Docker.  It allows you to extend a the docker image and add an `nginx.conf.tmpl` which will be dynamically parsed into the final configuration file based on state changes within the cluster.

This is a similar solution to `marathon-lb` but for the HTTP layer offering NGINX's advance http level uri routing.

### Overview

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
* Easy to get started add a `FROM containx/beethoven` to your `Dockerfile` add your template, config options and deploy!

### Getting Started

Below we will cover the barebones setup to get going.  

#### Create a Template

Create a file called `nginx.template`.  Refer to the `nginx.template` found in the [examples/](https://github.com/ContainX/beethoven/tree/master/examples) directory in this repo. Modify the example to suit your own needs 

**A couple notes about the example template:**  
- The `{{#if}}` blocks are optional.  I prefer these so if an application is removed all together in the cluster then the final `nginx.conf` is valid
- The `/_health` endpoint at the bottom is optional.  If allows for Marathon health checks to use that to determine Nginx is running. 
- The `/_bt` endpoint at the bottom is optional.  If you would like to find information such as updated times and any failures from Beethoven then this mapping allows you to expose these internal endpoints via Nginx.  Alternatively you can expose Beethoven via it's configured port.

### Create the Beethoven Configuration File

Create a file that ends in `.json`.  In this example we'll call it `btconf.json`. Refer to the `btconf.json` found in the [examples/](https://github.com/ContainX/beethoven/tree/master/examples) directory in this repo.
  
Add/Modify any options to suit your needs.  For a description and all possible configuration options refer to the docs found within the [config.go](https://github.com/ContainX/beethoven/blob/master/config/config.go) file.
 
### Create a Dockerfile

Next we will create the `Dockerfile` to package up the `nginx.template` and `btconf.json` files.  If you used the filenames in this guide then simply copy the code below into your `Dockerfile`.

```
FROM containx/beethoven

ADD nginx.template /etc/nginx/nginx.template
ADD btconf.json /etc/btconf.json
```

### Build and Testing your Container

Build and Run your Container

```
docker build -t myloadbalancer .
docker run -p 80:80 -d myloadbalancer -c /etc/btconf.json
```

Now open your browser and test paths you created at http://localhost


### Using a remote configuration file on Startup

A good practice is to keep configuration outside of Docker so the same container can be used between environments (QA, Prod, etc).  

Beethoven has built in support for using [spring-cloud-config](http://cloud.spring.io/spring-cloud-static/spring-cloud-config/1.2.0.RELEASE/) for centralized configuration.  Lets take the Docker example above but run it using our remote server.


```
docker run -p 80:80 -d myloadbalancer --remote --server http://spring-cloud-host:8888 --name myloadbalancer --profile prod
```

You can also specify the `--label` option which is the SCM branch the configuration server is pulling from. The names used above `profile, label and name` are the same names referenced in the official guide for `spring-cloud-config`. http://cloud.spring.io/spring-cloud-static/spring-cloud-config/1.2.0.RELEASE/


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