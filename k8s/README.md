# K8s Deployment

This assumes you have a K8s cluster with a Smart Agent running as a daemonset and configured to receive traces.  The complete-demo.yaml file is a K8s deployment for the complete Sock Shop app.  Each service will send traces to http://localhost:9080/v1/trace.

It also assumes you have a k8s namespace created called "sock-shop".

The provided configmap.yaml file is configured for SA 5.x using the new host monitors.  It is also configured to discover/monitor Mongo and RabbitMQ (which Sock Shop uses).  And it is configured for the new uAPM (not uAPM PG).  If you would like to test it with uAPM PG, you will need to spin-up a Smart Gateway somewhere (perhaps as a pod) and reconfigure the configmap.
