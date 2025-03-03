package helper

import (
	"github.com/go-resty/resty/v2"
)

func SendMail(email, subject, body string) error {
	emailName := GodotEnv("MAIL_ADDRESS")

	if email != "" {
		emailName = email
	}

	client := resty.New()
	_, err := client.R().
		SetBody(map[string]interface{}{
			"subject":    subject,
			"body":       body,
			"email":      emailName,
			"name":       "Bapak/Ibu yang terhormat",
			"email_type": 1,
		}).
		SetHeader("Content-Type", "application/json").
		SetAuthToken("Bearer " + GodotEnv("MAIL_TOKEN")).
		Post(GodotEnv("MAIL_HOST") + "authcorestaging/api/sendmailstatic")

	if err != nil {
		return err
	}
	return nil
}
