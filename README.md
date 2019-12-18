# sock-shop
This provides everything needed to run the Weaveworks Sock Shop demo app (https://microservices-demo.github.io/) in environments not otherwise provided by Weaveworks, starting with ECS/Fargate.  I plan to do EKS/Fargate next.  It is delivered as a CloudFormation stack that provisions all AWS resources necessary to run the app and monitor it with SignalFx instrumentation.  

You will incur a small cost to run the app; should be about $1/hour (in us-east-1 anyhow).

Currently, the application is instrumented with Smart Agent only, not uAPM.  I plan to do uAPM after EKS.  If others would like to work on the uAPM piece, that's fine; otherwise, I will get to it when I get to it.  I'd also like to work on a Terraform configuration in place of CloudFormation.  And I'd like to build this for Azure and GCP also.  I'll be busy!



INSTALLATION (FOR ECS-FARGATE)

See the how-to videos at https://signalfuse.atlassian.net/wiki/spaces/SE/pages/907936295/Sock+Shop+Fargate+Lab.

Requirements:
- You must have an AWS account of course! 
- A Key Pair configured in the EC2 console (easy to do).
- A SecureString parameter configured in the Systems Manager Parameter Store called 'sfx_access_token' that contains the SignalFx access token you wish to use.

The stack should work in the following regions (it will definitely fail in any other region):
us-east-1, us-east-2, us-west-1, us-west-2, eu-west-1, eu-central-1, ap-northeast-1, ap-southeast-1, ap-southeast-2

The region limitation is tied to an AMI that Weaveworks created and hosts in these regions.  

With those pieces in place, download the cfn-stack.yaml file for the deployment type you want (currently only ecs-fargate is available) and use it to create a stack in CloudFormation in the AWS Console.  You will need to give it a name and you will be prompted to select the Key Pair you want to use.  Then just click Next with defaults until you get to the screen with the orange "Create stack" button in the bottom right corner.

At this point, simply check the checkbox for "I acknowledge that AWS CloudFormation might create IAM resources."  You must check that box for the stack to run.  The stack does indeed create IAM resources.  Once checked, click the orange "Create stack" button.

In N. Virginia (us-east-1), it takes about 15 minutes to complete.  When you click "Create stack", it should take you to the Events tab which you can refresh and track progress.  Please report any errors you see.  I have tested in us-east-1 only (I see no reason why it shouldn't work in the other regions above though).

When it completes, click the Outputs tab at the top.  You will see the Application URL and Zipkin URL.  The Application URL is for the Sock Shop application.  Just click it; it's that easy!   The Zipkin URL is for an included Zipkin server.  A few of the Sock Shop services are configured to generate traces and send them to a Zipkin collector.  I kept all that for now for the curious.  As I implement uAPM, I will remove the Zipkin pieces -- but this can be done in a separate stack with different Docker images.


If you are studying for your AWS cert, you would do well to study the CloudFormation stack.  It creates and integrates many AWS resources like a VPC, subnets, IAM roles, Route 53, Internet Gateway, Security Groups, DynamoDB and RDS databases, EC2, ALB, Target Groups, CloudWatch logging, and a slew of Fargate tasks.  It touches a lot of material on the test.


REMOVAL

Just delete the CloudFormation stack.  It will delete everything it created as though you never ran it.  It takes a good 10-15 minutes to complete.  And you can track progress in the Events tab.

For thoroughness, when it is done, go to S3 and delete the "cfn-template*" bucket you will see there.  CloudFormation creates that to store stack templates that you upload, and you will incur a small charge for it if you don't delete it.  Otherwise, there should be no remnants.


USAGE NOTES

I use Visual Studio Code to read/edit the CloudFormation stack because it can roll-up sections of the file making it much easier to find sections you care about (like the individual task definitions that contain the Smart Agent sidecars), although other editors may do this also.  You can also look at the configurations for everything in the AWS Console after the stack is created. 

The stack uses a standard Internet Gateway (not NAT), assigns public IPs to the Fargate services, and secures them with Security Groups so that only the ALB is accessible from the internet.  The ALB in turn accesses the Fargate services in the VPC.  I considered using a NAT gateway and provide only private IPs to the Fargate services behind the ALB, but this requires two VPCs, a bastion host for internal access, and still requires an Internet Gateway, which complicates the config and it costs more.  So in the interest of minimizing cost and keeping the config reasonably simple, this is how I chose to do it.

All Smart Agent sidecars are configured to pull the agent.yaml file from this Github repository (in the ecs-fargate directory, for example).  All Fargate tasks are configured to send logs to CloudWatch in a Log Group named with the stack name that you provide when creating the stack.  You will likely spend a lot of time in CloudWatch reading the logs.  When you delete the stack, the CloudWatch logs are also deleted.

The stack spins up one EC2 that is used to load data into the RDS database.  This is why you need a Key Pair when running the stack.  The EC2 is a t2.micro that qualifies for the free tier if your account is still less than a year old.  The VPC Security Group allows you to SSH into this host.   Thus, the EC2 is useful as a bastion host to query the Smart Agents using the internal metrics URL (if you need to do that, which I did a few times).  ssh into the EC2 then run your queries from there.  Just go to EC2 in the Console and get the public IP.

If you have questions about anything in the CloudFormation stack, just ask.  

INT-1701 is open for an issue with our Redis integration showing recurring errors every 10 seconds in the Agent log that are not affecting operation but are annoying because they are spamming the log and are in fact invalid errors.   You will see those errors with Sock Shop (until it is fixed anyhow).

INT-1702 is open for an issue with our MongoDB integration showing a recurring error every 10 seconds in the Agent log that is affecting operation.  We are missing many default metrics.  This is happening specifically with the carts-db service of Sock Shop.  The user-db service uses an older version of Mongo and does not exhibit the error, so it is easy to suspect the issue is with the new version of Mongo that carts-db uses.  We'll see.


SETTING THE HOSTNAME WITH FARGATE

Every Fargate task (with the Smart Agent sidecar) will count as a host to SignalFx.  Each task is given a hostname and appears in Infrastructure Navigator like any other EC2 (even though they are not EC2s).  You can't tell the difference between a Fargate task and an EC2 in our UI.  In fact, the default hostname for a Fargate task takes the form of "ip-xxx-xxx-xxx-xxx.ec2.internal" so they are nearly impossible to identify in our UI unless you have very little else running.

This leads to the question of setting the hostname used for Fargate tasks so that you can identify them easily in our UI.

Options that I explored:
- CloudFormation offers a "Hostname" key under ContainerDefinitions (in the task definition).  This seemed like a simple no-brainer until I tried it and found that it is not supported with the awsvpc network mode used with Fargate.  Iow, it only works for EC2 launch types that do not use awsvpc.
- ECS provides a container metadata endpoint much like the EC2 metadata endpoint that can be queried with curl .  That metadata contains the task family name which would be perfect as the hostname for our purposes.  The challenge is that the family name is returned as part of a larger JSON blob, so we must then parse it out of the response.  Our Smart Agent Docker image does not provide utilities (like jq) for that.  So we'd have to go through the extra effort of installing those tools and then we get into rebuilding the Docker image, etc.  So this rightly becomes a feedback/enhancement request and not something we should hack (even though we could for a given customer).

For the above reasons, the simplest solution for now and for a POC is to set an environment variable in the Fargate task definition for the hostname we'd like to use and reference that environment variable in agent.yaml.  This requires setting the environment variable for each task definition.  And I chose to use variable SFX_HOSTNAME for this purpose.  You will see this variable in each task definition (in the CloudFormation stack) and in the agent.yaml file.

Finally, I looked for a way to reference the task family name when setting SFX_HOSTNAME in the task definition, but that does not work because the task must be created before you can reference the property using Fn::GetAtt (and since setting the variable is part of the task definition itself, the task obviously won't exist yet).  I believe that is the reason anyhow.  In any case, CloudFormation throws an error when you attempt it.  So we are left with hardcoding the value for SFX_HOSTNAME in the task definition.  Ack!

This is still not ideal though because we want to tie the hostname to a particular running instance of the task, so we can detect changes.  So I came up with a hack to achieve that, that you will see in the CloudFormation stack.  I export the SFX_HOSTNAME variable as part of the Command directive.  I placed it there instead of under Environment because it references another environment variable -- which is not allowed in the Environment section.  One gotcha after another!

I opened FEED-2477 to recommend we use ECS metadata (specifically the task family name combined with the task ID) to set the hostname that we use in SignalFx for Fargate tasks.


SHELL ACCESS

With K8s and Docker, you can shell into containers to troubleshoot issues and look at the environment.  You can't do that with Fargate!  It is possible only by adding sshd into the container, enabling it for root, configuring a key pair, and opening network access to SSH, all of which is a horrible idea for security and neither quick nor easy to do anyway.  Iow, you temporarily drop-in a special Docker image with root-enabled sshd baked in, then switch back when you're done.  Yuck.

Other options:
- Spin-up the task with an EC2 launch type.  This gives you CLI access to Docker then.  But you have to spin-up an EC2, join it to the cluster, temporarily reconfigure the task and service, etc.   Grrr..
- Spin-up the Docker image on your laptop (using Docker).  This is not helpful though if you need to see what is happening specifically with Fargate.  For instance, I wanted to see what environment variables Fargate sets, which we won't see this way.
- Use the 'Command' or 'EntryPoint' options in the task definition to override those options in the image and view the results in the CloudWatch log group.  For instance, if the original entrypoint in the Dockerfile is "java file.jar", you could modify the task definition to override that and instead run, for example, "printenv && java file.jar".  This outputs the environment variables then runs the java command after.  And this can be done without rebuilding the image so it is relatively easy.
- On second thought, this is easy in a lab environment.  It's not easy in prod because updating the task definition with a new Command and running it implies tearing down the old task and spinning up a new one, which implies impacting the service.

One helpful tip I picked up for checking on the Smart Agent without a shell is to add this line to agent.yaml (which you will see in this project's agent.yaml file):

internalStatusHost: ”0.0.0.0"

This allows us to hit the internal metrics URL from outside the container (on the EC2!) with:

curl http://<ip_address>:8095/?section=\<keyword\>

You can get the IP address out of the CloudWatch log.


INTERNAL DNS/SERVICE DISCOVERY

K8s handles internal DNS and service discovery (OOTB) much better than Fargate.  It is more or less automatic with K8s, at least with basic apps.  It can be done with Fargate, but it is not automatic even for basic apps.  You need to configure Service Discovery with each Fargate service which then triggers a combo of AWS Cloud Map and Route 53 configs.  And for the DNS resolution to work, you need to update your DHCP options (in the VPC) with the right domain -- oh, and make sure that DNS Resolution and DNS Hostnames are enabled for your VPC, which is not the default if the VPC is created via CloudFormation (know that one for the test!).  It's quite a pain, especially the more services there are.  Granted, there may be other service discovery solutions that are easier (Hashicorp's Consul?).  And Terraform (vs CloudFormation) may make it easier also.  I'm exploring this.

Where Fargate has it on K8s (and the ECS/EC2 launch type) is in the automated provisioning of nodes.  That is worth a lot!  You define your services and tasks; AWS does the rest.  It is very easy in comparison -- once you get internal DNS sorted out.  This is all done in the provided CloudFormation stack and you can look at the results in the Console.

And you may have heard the re:Invent announcement this year that AWS is now supporting Fargate with EKS.  I am betting that will be popular and eclipse ECS in time.
