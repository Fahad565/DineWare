package worker

import (
	"context"
	"fmt"
	"io"
	"log"
	"testing"
	"time"
)

type dispatcherTestRepository struct {
	stop     PendingStop
	notified bool
	queries  int
	marks    int
}

func (r *dispatcherTestRepository) QueryPendingStops(_ context.Context, _ time.Duration) ([]PendingStop, error) {
	r.queries++
	if r.notified {
		return nil, nil
	}
	return []PendingStop{r.stop}, nil
}

func (r *dispatcherTestRepository) MarkNotificationSent(_ context.Context, stopID int) error {
	r.marks++
	if stopID != r.stop.StopID {
		return fmt.Errorf("unexpected stop ID %d", stopID)
	}
	r.notified = true
	return nil
}

type dispatcherTestSMSSender struct {
	calls int
}

func (s *dispatcherTestSMSSender) SendSMS(context.Context, string, string, string, int, string) error {
	s.calls++
	return nil
}

func TestDispatchDueStopsSendsEachStopOnlyOnce(t *testing.T) {
	repo := &dispatcherTestRepository{
		stop: PendingStop{
			StopID: 17, ItineraryID: "ITIN-9502775", GroupName: "Safari Group",
			ChefPhone: "+254700000001", LodgeName: "Mara Lodge",
			CheckInDate: time.Now().Add(24 * time.Hour).Format("2006-01-02"),
			GuestCount:  2,
		},
	}
	sms := &dispatcherTestSMSSender{}
	dispatcher := NewDispatcher(repo, sms, log.New(io.Discard, "", 0))

	for cycle := 1; cycle <= 2; cycle++ {
		if err := dispatcher.DispatchDueStops(context.Background()); err != nil {
			t.Fatalf("dispatch cycle %d: %v", cycle, err)
		}
	}

	if sms.calls != 1 {
		t.Errorf("expected exactly one SMS across two cycles, got %d", sms.calls)
	}
	if repo.queries != 2 {
		t.Errorf("expected the repository to be queried twice, got %d", repo.queries)
	}
	if repo.marks != 1 {
		t.Errorf("expected stop notification to be marked once, got %d", repo.marks)
	}
}
