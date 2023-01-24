![Golang mascot](./gopher.png)

## About

This is simple library for control life cycle of your application.It provides you with extension points where you 
need to
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
go get github.com/andyollylarkin/go-app-lifecycle
```

## Usage

1. In the main function you need to create a new application instance.
2. Register your services
3. Call app.Run()

```go
func main() {
app := application.CreateApplication(zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger(), time.Second*3,
time.Second*9,
appMain, func () error { println("call user recover"); time.Sleep(time.Second * 10); return nil })

srv := new(DummyService)
app.RegisterService("dummy_example", srv)

err := app.Run()
if err != nil {
println("ERROR")
}
}
```

## Main function wait

```go
func appMain(ctx context.Context, wait func (), keeper *application.ServiceKeeper) error {
go func () {
for {
time.Sleep(time.Second * 2)
println("Main execution")
}
}()

wait() // Your application will block here and will wait until the end signal
// After unlocking, you will have N seconds to gracefully terminate your application
for {
time.Sleep(time.Second * 1)
println("SHUTDOWN")
}
return nil
}
```

## Services example

Service is an abstraction of the application life cycle. Anything that needs to be monitored, such as a database, can be
considered a service. To implement a service, you need to implement Service interface.

## TODO

- [ ] Remove zerolog from dependencies