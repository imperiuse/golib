package email

import "net/mail"

func ExampleMailBean_SendEmails() {
	Email := MailBean{}
	Email.Credentials = Credentials{"username", "password", "smtp.test.com:465", ""}
	Email.Subj = "Test Email Subj"
	Email.EnableNotify = true
	Email.SetFromAndToEmailAddresses(mail.Address{Name: "I", Address: "from@test.com"}, []mail.Address{{Name: "YOU", Address: "to@test.com"}})
	_ = Email.SendEmails("Body text of Email")
}
