# carvel-weather-app

This is Matt Royal's implementation of the Carvel Take Home Assessment.

## Running the app
**Prerequisite** You must have a recent version of go installed on your machine to run this app. I ran it with go 1.16.4

To run the app, cd into this directory and run `go run .`. This will start the app listening on port 8080.
You can change the port by setting the `PORT` environment variable to a port of your choice (e.g. `PORT=9000 go run .`)

Once it's running the app will output which port its listening on and then begin to log incoming requests.
You can shut down the app with `CTRL-C`.

The app implements the following endpoints, as laid out in the assignment:
- GET /weather
- POST /weather
- DELETE /erase

## Testing
The app was test driven, mostly via the integration tests in `integration/`, but also contains some unit tests.

You can run all tests with either of the following commands:
- `go test ./...`
- `ginkgo -r`

NOTE: On macOS Big Sur, running the tests will trigger several popups confirming that you want to allow the application
      listen on its port. Unfortunately I was not able to find a way to avoid this pop up.

## Future work
If I were to continue working on this app, here are some of the areas I'd want to focus on:

- Introduce data persistence. I opted for in-memory storage for simplicity given the short time frame, but this
  is obviously not ideal. I chose the repository pattern to make it simpler to swap in different persistence mechanisms
  in the future.

- Improve validation. I chose a few validations for the incoming JSON message, but there are likely others to be
  added. If I added persistence, I'd also want to add validations in the repository to ensure the data conformed to any
  data schema (e.g. max string length)

- Improve error messages. The error messages to the user are inconsistent and unstructured. I'd like to make them all
  JSON encoded, for example.

- More testing. I used the integration tests to drive most of my implementation, which left some code without unit
  tests. I'd also like to move some error handling tests in the integration spec into a more unit-style or
  lower-level integration test of the http handler functions.
