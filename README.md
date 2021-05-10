# Go currency rate

This is a gRPC web service using [Mortar](https://github.com/go-masonry/mortar)
Its main purpose is to show how easy it is for one to create a gRPC web service combined with [Temporal](https://temporal.io/) orchestrator to achieve reliable service that depends on other external APIs.

## Plan of work

* on hourly base, connect to currency exchange api and fetch rates for predefined currencies
* store the fetched rates in DB (MongoDB)
* expose two endpoints:
  * get history of requested currency
  * get current rate of requested currency
* The hourly connection cron implementation will be done using Temporal.io cron workflow