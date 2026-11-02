package viewer

// Catalog writers run only on a cloned record under TimelineV3Operation's lock.
func (s *Store) timelineWriteCatalog(v *TimelineSession, value any, object string, result *TimelineOperationResult) error {
	created := false
	switch r := value.(type) {
	case *TimelineTicketCreateRequest:
		input := r.TimelineTicketInput
		input.ExternalScope, _ = CanonicalTimelineIdentityURL(input.ExternalScope)
		for _, t := range v.Tickets {
			if t.ExternalScope == input.ExternalScope && t.ExternalID == input.ExternalID {
				result.TicketID = t.TicketID
				result.Created = &created
				return nil
			}
		}
		id, err := timelineRandomID("ticket-")
		if err != nil {
			return err
		}
		created = true
		v.Tickets = append(v.Tickets, TimelineTicket{TicketID: id, TimelineTicketInput: input})
		result.TicketID = id
		result.Created = &created
	case *TimelineDeliveryCreateRequest:
		scope, _ := CanonicalTimelineIdentityURL(r.RepositoryURL)
		for _, d := range v.Deliveries {
			if d.RepositoryURL == scope && d.Kind == r.Kind && d.ExternalID == r.ExternalID {
				result.DeliveryID = d.DeliveryID
				result.Created = &created
				return nil
			}
		}
		id, err := timelineRandomID("delivery-")
		if err != nil {
			return err
		}
		created = true
		v.Deliveries = append(v.Deliveries, TimelineDelivery{DeliveryID: id, RepositoryURL: scope, ExternalID: r.ExternalID, TimelineDeliverySnapshot: r.TimelineDeliverySnapshot})
		result.DeliveryID = id
		result.Created = &created
	case *TimelineCandidateCreateRequest:
		for _, c := range v.Candidates {
			if c.TimelineCandidate == r.TimelineCandidate {
				result.CandidateID = c.CandidateID
				result.Created = &created
				return nil
			}
		}
		id, err := timelineRandomID("candidate-")
		if err != nil {
			return err
		}
		created = true
		v.Candidates = append(v.Candidates, TimelineCandidateRecord{CandidateID: id, TimelineCandidate: r.TimelineCandidate})
		result.CandidateID = id
		result.Created = &created
	case *TimelineTicketPatchRequest:
		result.TicketID = object
		for i := range v.Tickets {
			t := &v.Tickets[i]
			if t.TicketID != object {
				continue
			}
			before := *t
			if r.Key != nil {
				t.Key = *r.Key
			}
			if r.Title != nil {
				t.Title = *r.Title
			}
			if r.URL != nil {
				t.URL = *r.URL
			}
			if before != *t {
				return timelineInvalidateCatalog(v, object, false)
			}
		}
	case *TimelineDeliveryPatchRequest:
		result.DeliveryID = object
		for i := range v.Deliveries {
			d := &v.Deliveries[i]
			if d.DeliveryID != object {
				continue
			}
			before := *d
			if r.Label != nil {
				d.Label = *r.Label
			}
			if r.URL != nil {
				d.URL = *r.URL
			}
			if r.Status != nil {
				d.Status = *r.Status
			}
			if before != *d {
				return timelineInvalidateCatalog(v, object, true)
			}
		}
	}
	return nil
}

func timelineInvalidateCatalog(v *TimelineSession, id string, delivery bool) error {
	tasks := map[string]bool{}
	ticketLinked := func(refs []TimelineTicketRef) bool {
		for _, ref := range refs {
			if ref.TicketID == id {
				return true
			}
		}
		return false
	}
	for _, e := range v.Events {
		linked := !delivery && ticketLinked(e.TicketRefs)
		if delivery {
			for _, ref := range e.DeliveryRefs {
				linked = linked || ref.DeliveryID == id
			}
		}
		for _, view := range e.Views {
			linked = linked || !delivery && ticketLinked(view.TicketRefs)
		}
		if linked {
			for _, task := range timelineRelatedTasks(*v, e.EventID) {
				tasks[task] = true
			}
			if e.TaskID != "" {
				tasks[e.TaskID] = true
			}
		}
	}
	if delivery {
		for i := range v.Waits {
			w := &v.Waits[i]
			if w.PullRequestRef.DeliveryID == id {
				if w.WaitVersion >= TimelineMaxNumber {
					return timelineFailure("capacity", 413)
				}
				w.WaitVersion++
				tasks[w.TaskID] = true
			}
		}
	}
	ids := make([]string, 0, len(tasks))
	for task := range tasks {
		ids = append(ids, task)
	}
	return timelineBumpTask(v, ids...)
}
