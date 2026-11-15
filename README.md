# go-bancontactpro

A collection of Go clients for the Bancontact Pro merchant APIs. See [Bancontact Pro developer portal](https://docs.bancontactpro.com/)

## Scope

Currently, only a subset of the [Payment API](https://docs.bancontactpro.com/apis/merchant-payment.openapi) is supported using API keys.


## Usage

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/oeaniz/go-bancontactpro/payments"
)

func main() {
    // Store the API Key in a secret manager, not in source control.
    apiKey := os.Getenv("BANCONTACT_API_KEY")

    client, err := payments.NewClient(apiKey, payments.WithEnvironment(payments.EnvironmentPreProduction))
    if err != nil {
        log.Fatal(err)
    }

    payment, err := client.CreatePayment(context.Background(), payments.CreatePaymentRequest{
        Amount:   1250, // EUR cents
        Currency: payments.CurrencyEUR,
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("created payment: %s", payment.PaymentID)
    if payment.Links.QRCode != nil {
        log.Printf("url of QR code for payment: %s", payment.Links.QRCode.Href)
    }
}
```

`WithBaseURL` can direct the client to a compatible mock server. The optional integration test suite exercises the published mock API:

```sh
go test -tags=integration ./payments
```

Set `BANCONTACT_MOCK_URL` to override the default mock API URL.
