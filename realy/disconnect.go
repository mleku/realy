package realy

func (s *Server) Disconnect() {
	for client := range s.clients {
		log.I.F("closing client %s", client.RemoteAddr())
		client.Close()
	}
}
