# catalogue service

**SignalFx uAPM Instrumentation**

<br/>

For starters, the good folks at Weaveworks delivered the Go-based services (of which catalogue is one) already manually-instrumented with OpenTracing and OpenZipkin -- and specifically for calls to the net/http library (and not the other libraries).

As delivered by Weaveworks, the traces were sent using an old version of Thrift that we (SignalFx) do not support.  We kicked back errors on it.  So because it was quick and easy, I first changed the delivery format to Zipkin.  But that had issues that I attribute to bugs with OpenZipkin and Go.  Rather than spend time debugging that, I switched to using SignalFx's tracer.  For time's sake, I did not leverage the auto-instrument capabilities of our agent.  I was under a deadline to deliver a demo, so this represents technical debt.

I added SignalFx's tracer in a way that still leverages the OpenTracing calls that are already there.  Which is to say, this is not by the book, but it may still be instructive to folks to see what I did.

I modified two files primarily:

- ./vendor/manifest

This file contains the dependencies for the app that are needed to compile.  I added the signalfx tracer.  Look for these lines:
{
	"importpath": "github.com/signalfx/signalfx-go-tracing",
	"repository": "https://github.com/signalfx/signalfx-go-tracing",
	"vcs": "git",
	"branch": "master",
	"notests": true
},

I deleted stuff related to OpenZipkin.  The original manifest file is also there to compare. For better or worse, these manifest files are part of an obsolete Go dependency mgmt tool called gvt that you are unlikely to still see in use.  That said, I added the above lines and it worked fine.

Adding our tracer as a dependency is something you will most likely need to do for any Go project.  So it becomes a question of how to do that, and a frustrating aspect of Go is that there have been *many* quite different dependency mgmt tools over the years.  So that part may require some figuring out per project.

<br/>

- ./cmd/cataloguesvc/main.go

This is the main catalogue app.  The original version is there for you to compare.  The SignalFx tracer is implemented in a way that differs greatly from the documentation; however, it will use env variables for the service name and trace endpoint URL -- and I always prefer to use env variables for that, especially with Go, so that you can change those variables without rebuilding the app.

A consequence of how I implemented the tracer is that it does not automatically use B3 headers to associate upstream and downstream spans.  You need to set these env variables for that to work:
name: DD_PROPAGATION_STYLE_INJECT
value: "B3"
name: DD_PROPAGATION_STYLE_EXTRACT
value: "B3"

You can do this in the Fargate task definition or in a K8s deployment spec.  You will see I did this already in the CloudFormation template.

My technical debt is to go back and do this in a more "by the book" way that auto-instruments both net/http and jmoiron/sqlx.

Bear in mind that it is a stretch to call this "auto-instrumentation" even when we do it by the book.  That's not our fault though.  Go is a compiled language.  We make it as easy as it can be, but it still involves significant changes to code and compiling it into the app.  No way around that.

Thankfully, Weaveworks compiles the app as part of the Dockerfile so that we do not need to do that as a separate step, and you don't need to install Go on your laptop.

<br/>

Finally, I had to change the Dockerfile to use the "golang:latest" image (very first line of the file).  It originally used go:1.7, which is very old now and the SignalFx tracer does not work with it (we require 1.12+).  Thankfully the application worked ok with the latest version with just a few dependency and syntax adjustments in the code (which I made and you won't have to worry about).
