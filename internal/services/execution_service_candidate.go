package services

import (
	"context"
	"fmt"
)

// candidateKey namespaces the in-memory/persisted session registry so a
// candidate session never collides with appID's own plain (active) key.
// This is the only place that composition happens — StartCandidate,
// CandidateSession and PromoteCandidate all go through it, so the format
// can never drift between them.
func candidateKey(appID string) string {
	return appID + ":candidate"
}

func (s *executionService) StartCandidate(ctx context.Context, appID string, port int) (RunSession, error) {
	app, err := s.repo.FindByID(ctx, appID)
	if err != nil {
		return RunSession{}, err
	}

	strategy, err := s.resolver.Resolve(app)
	if err != nil {
		return RunSession{}, err
	}

	session, err := strategy.Start(ctx, app, port)
	if err != nil {
		return RunSession{}, err
	}
	s.trackSession(ctx, candidateKey(appID), session)
	return toRunSession(session), nil
}

func (s *executionService) CandidateSession(appID string) (RunSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[candidateKey(appID)]
	if !ok {
		return RunSession{}, false
	}
	return toRunSession(session), true
}

func (s *executionService) StopCandidate(ctx context.Context, appID string, session RunSession) error {
	app, err := s.repo.FindByID(ctx, appID)
	if err != nil {
		return err
	}

	strategy, err := s.resolver.Resolve(app)
	if err != nil {
		return err
	}

	if err := strategy.Stop(ctx, app, fromRunSession(session)); err != nil {
		return err
	}
	s.untrackSession(ctx, candidateKey(appID))
	return nil
}

func (s *executionService) PromoteCandidate(ctx context.Context, appID string) error {
	s.mu.Lock()
	candidate, ok := s.sessions[candidateKey(appID)]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("services: no candidate session tracked for application %q", appID)
	}

	s.trackSession(ctx, appID, candidate)
	s.untrackSession(ctx, candidateKey(appID))

	if app, err := s.repo.FindByID(ctx, appID); err == nil {
		s.syncApplicationStatus(ctx, app, candidate.Status)
	}
	return nil
}

func (s *executionService) CheckStatus(ctx context.Context, appID string, session RunSession) (RunSession, error) {
	app, err := s.repo.FindByID(ctx, appID)
	if err != nil {
		return RunSession{}, err
	}

	strategy, err := s.resolver.Resolve(app)
	if err != nil {
		return RunSession{}, err
	}

	updated, err := strategy.Status(ctx, app, fromRunSession(session))
	if err != nil {
		return RunSession{}, err
	}
	return toRunSession(updated), nil
}

func (s *executionService) StopSession(ctx context.Context, appID string, session RunSession) error {
	app, err := s.repo.FindByID(ctx, appID)
	if err != nil {
		return err
	}

	strategy, err := s.resolver.Resolve(app)
	if err != nil {
		return err
	}

	return strategy.Stop(ctx, app, fromRunSession(session))
}
