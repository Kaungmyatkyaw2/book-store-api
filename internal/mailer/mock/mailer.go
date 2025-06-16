package mock

type Mailer struct {
}

func (m Mailer) Send(receipent, templateFile string, data interface{}) error {
	return nil
}
