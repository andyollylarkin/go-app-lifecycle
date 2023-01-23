![Golang mascot](https://listimg.pinclipart.com/picdir/s/571-5718150_go-clipart.png)

## About

This is simple library for control lifecycle of your application.It provides you with extension points where you need to
put your own main application code and controls the health of "services" in the background. In the event of an error in
the main application or failure of the "service", this will be intercepted by the library and you will be given time to
gracefully complete the main work (e.g. [http.Server.Shutdown()](https://pkg.go.dev/net/http#Server.Shutdown)).

#### App lifecycle:

1. Init phase - start your services initialization (e.g. establish db connection). Then runs in the background
   services health checking.
2. Start phase - start your main application function.
3. Shutdown phase - in case of an error in the application, this phase is started. It gives the application a cutoff for
   gracefully shutting down its work.
4. Uninitialization phase - uninitialization services. E.g. close db connection.
5. Application terminates.

## Installation

```shell

```

## Usage

## Main function wait

## Getting services

## TODO