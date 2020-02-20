# user service

**SignalFx uAPM Instrumentation**

<br/>

For starters, the good folks at Weaveworks delivered the Go-based services (of which user is one) already manually-instrumented with OpenTracing and OpenZipkin -- and specifically for calls to the net/http library (and not the other libraries).

As delivered by Weaveworks, the traces were sent using an old version of Thrift that we (SignalFx) do not support.  We kicked back errors on it.  So because it was quick and easy, I first changed the delivery format to Zipkin.  But that had issues that I attribute to bugs with OpenZipkin and Go.  Rather than spend time debugging that, I switched to using SignalFx's tracer.  For time's sake, I did not leverage the auto-instrument capabilities of our agent.  I was under a deadline to deliver a demo, so this represents technical debt.

I added SignalFx's tracer in a way that still leverages the OpenTracing calls that are already there.  Which is to say, this is not by the book, but it may still be instructive to folks to see what I did.

I modified these files:

The other Go services in Sock Shop use a dependency mgmt tool called gvt.  This service (for whatever reason) uses a different also-obsolete dependency mgmt tool called glide.

- ./glide.lock and glide.yaml

These files contain the dependencies for the app that are needed to compile.  I added the signalfx tracer to both files.  I deleted stuff related to OpenZipkin.

Adding our tracer as a dependency is something you will most likely need to do for any Go project.  So it becomes a question of how to do that, and a frustrating aspect of Go is that there have been *many* quite different dependency mgmt tools over the years.  So that part may require some figuring out per project.

<br/>

- ./main.go

This is the main user app.  The original version is there for you to compare.  The SignalFx tracer is implemented in a way that differs greatly from the documentation; however, it will use env variables for the service name and trace endpoint URL -- and I always prefer to use env variables for that, especially with Go, so that you can change those variables without rebuilding the app.

A consequence of how I implemented the tracer is that it does not automatically use B3 headers to associate upstream and downstream spans.  You need to set these env variables for that to work:
name: DD_PROPAGATION_STYLE_INJECT
value: "B3"
name: DD_PROPAGATION_STYLE_EXTRACT
value: "B3"

You can do this in the Fargate task definition or in a K8s deployment spec.  You will see I did this already in the CloudFormation template.

My technical debt is to go back and do this in a more "by the book" way that auto-instruments net/http.

Bear in mind that it is a stretch to call this "auto-instrumentation" even when we do it by the book.  That's not our fault though.  Go is a compiled language.  We make it as easy as it can be, but it still involves significant changes to code and compiling it into the app.  No way around that.

Thankfully, Weaveworks compiles the app as part of the Dockerfile so that we do not need to do that as a separate step, and you don't need to install Go on your laptop.

<br/>

Finally, I had to change the Dockerfile to use the "golang:1.13.7-alpine" image (very first line of the file).  It originally used go:1.7-alpine, which is very old now and the SignalFx tracer does not work with it (we require 1.12+).  

Unlike the other Go services, this upgrade did not go smoothly.  Many things broke.  I had to upgrade many dependencies.  Also the application itself broke.  I had to modify the ./users/links.go source file.  The changes are not dramatic, but it was tough figuring it out -- it has to do with how the json responses are formed.

However this was a happy accident.  I built Docker images for both the broken and working versions of the app.  The broken version fails when you try to place an order in the app.  It throws a 500 error and we can drilldown to the traces to troubleshoot.  So it makes for a good demo.  You can choose which image you'd like to use.  Both images are on Docker Hub under marksfink/sock-shop-user.  Look at the tags.
