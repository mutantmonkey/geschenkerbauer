package main

import (
	"fmt"
	"log"
	"net/http"
)

func triggerWebhook(config Config) error {
	if config.Webhook.URL != "" {
		log.Print("Trigger webhook")

		req, err := http.NewRequest("POST", config.Webhook.URL, nil)
		if err != nil {
			return err
		}

		if config.Webhook.BearerToken != "" {
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Webhook.BearerToken))
		}

		_, err = http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
	}

	return nil
}
