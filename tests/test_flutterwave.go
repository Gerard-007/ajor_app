package main

import (
    "fmt"
    "strings"
    "net/http"
    "io"
)

func main() {
    url := "https://api.flutterwave.com/v3/virtual-account-numbers"
    payload := strings.NewReader(`{"email":"user@example.com","currency":"NGN","amount":2000,"tx_ref":"jhn-mdkn-10192029920","is_permanent":false,"narration":"Please make a bank transfer to John","phonenumber":"08012345678"}`)

    req, _ := http.NewRequest("POST", url, payload)
    req.Header.Add("accept", "application/json")
    req.Header.Add("Authorization", "Bearer FLWSECK_TEST-e51cb7b74a4b38998fb13148d8ef0095-X") // Use your test key
    req.Header.Add("Content-Type", "application/json")

    res, err := http.DefaultClient.Do(req)
    if err != nil {
        fmt.Println("Request error:", err)
        return
    }

    defer res.Body.Close()
    body, err := io.ReadAll(res.Body)
    if err != nil {
        fmt.Println("Read error:", err)
        return
    }

    fmt.Println("Response:", string(body))
}