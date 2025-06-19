package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
    url := "https://api.flutterwave.com/v3/virtual-account-numbers"
    payload := strings.NewReader(`{"email":"user@example.com","currency":"NGN","amount":2000,"tx_ref":"jhn-mdkn-10192029920","is_permanent":false,"narration":"Please make a bank transfer to John","phonenumber":"08012345678"}`)

    req, _ := http.NewRequest("POST", url, payload)
    req.Header.Add("accept", "application/json")
    req.Header.Add("Authorization", "Bearer " + os.Getenv("FLW_SECRET_KEY")) // Use your test key
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
