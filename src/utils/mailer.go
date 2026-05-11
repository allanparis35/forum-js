package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendResetEmail(toEmail string, token string) error {
	//Récupération des infos du serveur SMTP depuis le .env
	from := os.Getenv("MAILER_USER")
	password := os.Getenv("MAILER_PASSWORD")
	host := os.Getenv("MAILER_HOST") 
	port := os.Getenv("MAILER_PORT")

	//Construction de l'URL que l'utilisateur cliquera dans son mail 
	resetLink := fmt.Sprintf("http://localhost:5173/reset-password?token=%s", token)

	//Création du mail
	subject := "Subject: Reinitialisation de votre mot de passe\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<html>
			<body>
				<h2>Bonjour,</h2>
				<p>Vous avez demandé à réinitialiser votre mot de passe sur le Forum.</p>
				<p>Cliquez sur le lien ci-dessous (valable 15 minutes) :</p>
				<a href="%s" style="padding: 10px 20px; background-color: #007bff; color: white; text-decoration: none; border-radius: 5px;">
					Réinitialiser mon mot de passe
				</a>
				<p>Si vous n'êtes pas à l'origine de cette demande, ignorez cet email.</p>
			</body>
		</html>
	`, resetLink)

	msg := []byte(subject + mime + body)
	auth := smtp.PlainAuth("", from, password, host)

	err := smtp.SendMail(host+":"+port, auth, from, []string{toEmail}, msg)
	if err != nil {
		return err
	}

	return nil
}