package core


type Service struct {
	Engine Engineer
	Server *Server
}


func (s *Service) Run()  {
	// Engine run
	s.Engine.Run()
	// Server run
	s.Server.Run()
}



