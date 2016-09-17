# Beethoven

Is a Mesos/Marathon stream listener which manages an NGINX proxy instance within Docker.  It allows you to extend a the docker image and add an `nginx.conf.tmpl` which will be dynamically parsed into the final configuration file based on state changes within the cluster.

This is a similar solution to `marathon-lb` but for the HTTP layer offering NGINX's advance http level uri routing.

This is currently a WIP -- documentation will be updated when the initial development is complete