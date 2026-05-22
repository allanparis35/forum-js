package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func Send2FACodeEmail(toEmail string, code string) error {
	//Récupération des infos du serveur SMTP depuis le .env
	from := os.Getenv("MAILER_USER")
	password := os.Getenv("MAILER_PASSWORD")
	host := os.Getenv("MAILER_HOST") 
	port := os.Getenv("MAILER_PORT")
	//Création du mail
	subject := "Subject: Votre code de vérification 2FA\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<html>
			<body>
				<h2>Bonjour,</h2>	
				<p>Voici votre code de vérification pour finaliser votre inscription sur le Forum :</p>
				<h1 style="color: #007bff;">%s</h1>
				<p>Ce code est valable 15 minutes.</p>	
				<p>Si vous n'êtes pas à l'origine de cette demande, ignorez cet email.</p>
			</body>
		</html>
	`, code)	

	msg := []byte(subject + mime + body)
	auth := smtp.PlainAuth("", from, password, host)
	err := smtp.SendMail(host+":"+port, auth, from, []string{toEmail}, msg)
	if err != nil {
		return err
	}
	return nil
}