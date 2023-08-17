package salesforce

func (s *Connector) Close() error {
	s.Client.CloseIdleConnections()

	return nil
}
