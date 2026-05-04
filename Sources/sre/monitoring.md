# Monitoring 

Some thoughts on monitoring.

Source Documents:
+ https://docs.google.com/document/d/1x-frPXiGRXeJp8f-asp0LVdSoM1IAXEKSlPuYvD30fw/edit
+ https://www.bigpanda.io/resource-library/monitoringscape/
+ https://engineering.salesforce.com/monitoring-microservices-divide-and-conquer-acca62b209cc
+ https://www.circonus.com/2018/06/comprehensive-container-based-service-monitoring-with-kubernetes-and-istio/
+ https://www.vividcortex.com/blog/monitoring-and-observability-with-use-and-red

### Organization Metrics:
From DevOps Research and Assessment (DORA) and ["Accelerate: The Science of Lean Software and DevOps"](https://itrevolution.com/book/accelerate/):
+ **Deployment Frequency**
+ **Lead Time for Changes**
+ **MTTR - Mean Time To Resolution**
+ **Change Failure Rate**

The thorough [State of DevOps](https://devops-research.com/research.html) reports have focused on data-driven and statistical analysis of high-performing organizations. The result of this multiyear research, published in [Accelerate](https://itrevolution.com/book/accelerate/), demonstrates a direct link between organizational performance and software delivery performance. The researchers have determined that only four key metrics differentiate between low, medium and high performers: lead time, deployment frequency, mean time to restore (MTTR) and change fail percentage. Indeed, we've found that these four key metrics are a simple and yet powerful tool to help leaders and teams focus on measuring and improving what matters. A good place to start is to instrument the build pipelines so you can capture the four key metrics and make the software delivery value stream visible. 

### Team Metrics:
+ Story Counting = Number of items in the Backlog
+ Lead Time = Average Time From Backlog to Done
+ Cycle Time = Average Time From Started to Done
+ Wait/Queue Time = Lead Time - Cycle Time

### SLIs drive SLOs, which inform SLAs.

A Service Level Indicator (SLI) is a metric derived measure of health for a service. For example, I could have an SLI that says my 95th percentile latency of homepage requests over the last 5 minutes should be less than 300 milliseconds.

A [Service Level Objective](https://en.wikipedia.org/wiki/Service_level_objective) (SLO) is a goal or target for an SLI. We take an SLI, and extend its scope to quantify how we expect our service to perform over a strategic time interval. Using the SLI from the previous example, we could say that we want to meet the criteria set by that SLI for 99.9% of a trailing year window.

A [Service Level Agreement](https://en.wikipedia.org/wiki/Service-level_agreement) (SLA) is an agreement between a business and a customer, defining the consequences for failing to meet an SLO. Generally, the SLOs which your SLA is based upon will be more relaxed than your internal SLOs, because we want our internal facing targets to be more strict than our external facing targets.

We also recommend watching the YouTube video "[SLIs, SLOs, SLAs, oh my!](https://youtu.be/tEylFyxbDLE)" from Seth Vargo and Liz Fong Jones to get an in depth understanding of the difference between SLIs, SLOs, and SLAs.

+ https://roadmap.sh/guides/what-is-sli-slo-sla

### SLIs are RED
What SLIs best quantify host and service health? Over the past several years, there have been a number of emerging standards. The top standards are the [USE method](http://www.brendangregg.com/usemethod.html), the [RED method](https://www.weave.works/blog/the-red-method-key-metrics-for-microservices-architecture/), and the “[four golden signals](https://landing.google.com/sre/book/chapters/monitoring-distributed-systems.html#xref_monitoring_golden-signals)” discussed in the Google SRE book.

##### USE = utilization, saturation, errors
+ **U**tilization: the average time that the resource was busy servicing work
+ **S**aturation: the degree to which the resource has extra work which it can't service, often queued
+ **E**rrors: the count of error events

This disambiguates utilization and saturation, making it clear that utilization is "busy time %" and saturation is “backlog.” These terms are very different from things a person might confuse with them, such as “disk utilization” as an expression of how much disk space is left.

##### RED = Rate, Errors, and Duration
Tom Wilkie introduced the [RED method](https://www.weave.works/blog/the-red-method-key-metrics-for-microservices-architecture/) a few years ago. With RED we monitor *request rate*, *request errors*, and *request duration*. The Google SRE book talks about using latency, traffic, errors, and saturation metrics. These “[four golden signals](https://landing.google.com/sre/book/chapters/monitoring-distributed-systems.html#xref_monitoring_golden-signals)” are targeted at service health, and is similar to the RED method, but extends it with saturation. In practice, it can be difficult to quantify service saturation.

Again, the RED Method defines the three key metrics you should measure for every microservice in your architecture as:

+ (Request) **R**ate - the number of requests, per second, your services are serving.
+ (Request) **E**rrors - the number of failed requests per second.
+ (Request) **D**uration - distributions of the amount of time each request takes.


###### USE RED Together?

What may not be obvious is that USE and RED are complementary to one another. The USE method is an internal, service-centric view. The system or service’s workload is assumed, and USE directs attention to the resources that handle the workload. The goal is to understand how these resources are behaving in the presence of the load.

The RED method, on the other hand, is about the workload itself, and treats the service as a black box. It’s an externally-visible view of the behavior of the workload as serviced by the resources. Here workload is defined as a population of requests over a period of time. It is important to measure the workload, since the system’s raison d’etre is to do useful work.

Taken together, RED and USE comprise minimally complete, maximally useful observability—a way to understand both aspects of a system: its users/customers and the work they request, as well as its resources/components and how they react to the workload. (I include users in the system. Users aren’t separate from the system; they’re an inextricable part of it.)

+ U = Utilization, as canonically defined
+ S = Saturation - Measure Concurrency
+ E = Error Rate, as a throughput metric
+ R = Rate - Request Throughput, in requests per second
+ E = Error - Request Error Rate, as either a throughput metric or a fraction of overall throughput
+ D = Duration - Request Latency, Residence Time, or Response Time; all three are widely used

### SLOs

Once we define all the indicators and collect metrics for all of them, we then need to decide what is good and what is bad. To do this we should make 2 steps: baseline the metrics, decide what is acceptable for every metric and where the acceptable range ends.

With the numeric definition of the acceptable ranges we define Service Level Objectives (SLOs). You can read more about [SLOs on Wikipedia](https://en.wikipedia.org/wiki/Service_level_objective). Examples of SLOs are that service should have 99.9% availability over a year, or that the 95th percentile of latency for responses should be below 300ms over the course of a month. It’s always better to keep some buffer between the announced SLO and zones where things start going really badly.

### SLAs

A SLA is a two way agreement: clients of our service agree on conditions for using it, and we promise them that under those conditions the service will perform within some boundaries. 

Clients of the service want to know what can they expect from it in terms of performance and availability: how many requests per second it can process, length of expected downtime during maintenance, or how long it takes on average to process a request. Usually, performance and availability of a service can be expressed using very few parameters, and in most cases the list can be applied to other services also.

If request rate grows beyond the agreed level, we can start throttling requests or denying to serve requests (first communicating this action to the client, of course). If latency grows beyond declared limits and there is no significant increase of the request rate, then we know that something is wrong on our side and it’s time for us to begin troubleshooting.

### Divide and Conquer strategy

When a product is a network of interconnected services with a rich collection of external dependencies, it’s really hard to identify bottlenecks or unhealthy services. A single model for the whole system is too complex to define and understand.

A good strategy here is to divide and conquer. Monitoring can be simplified dramatically by focusing on every service separately and tracking how others use it and how it uses the services it depends on. This can be accomplished by following three simple rules:

1. Every service should have its own Service Level Agreement (SLA).
2. Every instance of every service monitors how others use it and how it responds.
3. Every instance of every service monitors how it uses other services and how they respond.

Of course every instance of every service has a health check and produces metrics about its internal state to ease troubleshooting. In other words, every instance of every service is a white box for owners of the service but it’s a black box for everyone else.

### From [Non-standard SLOs: beyond availability](https://isitobservable.io/events/non-standard-slos-beyond-availability)

What are SLOs?
--------------

An [SLO is a Service Level Objective](https://www.dynatrace.com/news/blog/what-are-slos/), meaning the level that you expect a service to achieve most of the time and against which an SLI (Service Level Indicator) is measured.

An SLO measures the ratio of good/total over a time duration. A human will be notified if it’s not good for too long.

[![SLO and SLI measurement](https://isitobservable.io//assets/images/defining-SLOs-graph.png)](https://isitobservable.io//assets/images/defining-SLOs-graph.png)

In most cases, SLOs would be used to measure the availability (is it up?) or the latency (is it fast?) of request-driven services. But what would you use as SLOs for things like data processing, scheduled execution, and ML models? Here are some ideas proposed by Steve.

Data processing - Freshness
---------------------------

_The proportion of valid data updated more recently than a threshold._

If you're generating map tiles for something like google maps, you may know that there is a pipeline system where information comes in on one side and exits on the other. The information will get a timestamp at the exit, for example, when it was generated. If we compare this timestamp to when it was served to the user, you can calculate how old or new the map tile is.

A good value can be defined as the delta between when it was built and when it was served compared to some threshold (an acceptable level of freshness).

```java
Good = datetime_served - datetime_built < threshold
```

Data processing - Coverage
--------------------------

_The proportion of valid data processed successfully._

If you have a lot of inputs going into different pipelines and going through different outputs, we want to make sure we understand the state of the whole system and see if it drops any inputs for various reasons. For example, is it ok if the system drops 90% of the things it should be processing? It’s up to you and your SLO to decide.

```java
Good = valid_records - num_processed < threshold
```

Scheduled Execution - Skew
--------------------------

_The time difference between when the job should have started and when it did._

Scheduled jobs don’t always run exactly at the time you schedule them. It’s either late, early, or on time. It’s useful to be flexible, but you need to know how it’s doing. You can measure the time between the scheduled execution and when it started. What is the max allowed threshold?

```java
Good = time_started - time_scheduled < max_thresholdandTime_started - time_scheduled > min_threshold
```

Sometimes, you may be okay with a threshold of 24 hours. In other cases, just 5 minutes.

Knowing what percentage of these executions fit into that window is really useful for understanding the big picture of what's going on in your system.

Scheduled Execution - Duration
------------------------------

_The time difference between when the job should have been completed and by when it was expected to complete._

In this SLO, you want to know for how long the execution ran. If a program usually runs for an hour, but you notice over time that the execution goes over that amount of time, you can add more resources to get it under that threshold.

```java
good = (time_ended or NOW) - time_started < max_expectedandtime_ended - time_started > min_expected
```

The duration helps you understand if a large system is fast enough as a whole. Setting expected upper/lower boundaries provides a method for knowing what is considered good. (Even jobs that are too short could be an issue.)

Durability
----------

The durability SLO can be confusing as it tends to have many nines (11), and you'd think at a glance that it's all good. However, durability measures a predicted distribution of potential physical failure modes over time. You can model, understand, and improve durability by adding physical replicas and encoding schemes for storage.

Durability is very different from latency and other SLOs. You can't compare them directly.

Conclusion: SLOs
----------------

SLOs are a measurement of an entire system, not just components. They're an abstraction of how well your system performs at a very high level. That means that you still need to do a deep diagnosis. They don’t replace monitoring but may replace alerting if the alerts are related to your customer's happiness.

With SLOs, you define certain terms for the whole company so everybody understands them the same way: service, goal, criteria, period, performance, period, error budget, etc. They help you interact more transparently without human interpretation.

And this type of abstraction provides a consistent understanding of behavior through change. If you introduce a new system, you can make direct comparisons to show how a new system is better than the old one, thanks to SLOs.