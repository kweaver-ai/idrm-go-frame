package corn

func (s *SyncMiddleware) processIncrSync(corn *CornConf) (any, error) {
	tx, err := s.db.Begin()

	if err != nil {
		return nil, err
	}
	return nil, tx.Commit()
}
