package email

import "net/mail"

func ExampleMailBean_SendEmails() {
	Credentials := Credentials{"username", "password", "smtp.test.com:465", ""}
	Email := NewMailBean(nil, mail.Address{Name: "I", Address: "from@test.com"}, []mail.Address{{Name: "YOU", Address: "to@test.com"}}, "Subj mail", Credentials, true)
	//Email.SetFromAndToEmailAddresses(mail.Address{Name: "I", Address: "from@test.com"}, []mail.Address{{Name: "YOU", Address: "to@test.com"}})
	_ = Email.SendEmails("Body text of Email")
}
