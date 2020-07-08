package factom-eater

type Eater struct {
	listener *net.Listener
	stopper chan bool
}

func Launch(host string) (*Eater, error) {
	listener, err := net.Listen(host)
	if err != nil { return nil, err }



	e := new(Eater)
	e.stopper = make(chan bool)
	e.listener = listener
	go e.listen()
	return e, nil
}

func (e *Eater) listen() {

}