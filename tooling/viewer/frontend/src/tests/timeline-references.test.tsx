import assert from "node:assert/strict";
import test from "node:test";
import { renderToStaticMarkup } from "react-dom/server";
import { completeFixture } from "../dev/timeline-complete-fixture";
import {
  resolveDelivery,
  resolveTicket,
  TimelineDeliveryLinks,
  TimelineTicketLinks,
} from "../timeline-references";

test("current references resolve opaque IDs while publication snapshots stay immutable", () => {
  const { session } = completeFixture();
  const ticket = session.tickets[0];
  const ref = {
    ticket_id: ticket.ticket_id,
    snapshot: { key: ticket.key, url: ticket.url },
  };
  const original = JSON.stringify(ref);
  ticket.key = "CORRECTED-7";
  ticket.url = "https://example.test/corrected/7";
  assert.equal(resolveTicket(session, ref), ticket);
  const html = renderToStaticMarkup(
    <TimelineTicketLinks session={session} refs={[ref]} lang="es" />,
  );
  assert.match(html, /CORRECTED-7/);
  assert.match(html, /https:\/\/example.test\/corrected\/7/);
  assert.match(html, /Publicado/);
  assert.match(html, /Actual/);
  assert.equal(JSON.stringify(ref), original);
  const deliveryRef = session.events.find((e) => e.delivery_refs?.length)
    ?.delivery_refs?.[0];
  assert.ok(deliveryRef);
  const delivery = resolveDelivery(session, deliveryRef);
  assert.ok(delivery);
  delivery.status = "merged";
  delivery.label = "PR corrected";
  delivery.url = "https://example.test/pull/updated";
  const deliveryHTML = renderToStaticMarkup(
    <TimelineDeliveryLinks session={session} refs={[deliveryRef]} lang="es" />,
  );
  assert.match(deliveryHTML, /Publicado/);
  assert.match(deliveryHTML, /Actual/);
  assert.match(deliveryHTML, /PR corrected/);
  assert.equal(deliveryRef.snapshot.status, "open");
});

test("unknown IDs never impersonate a current entity using captured values", () => {
  const { session } = completeFixture();
  const ref = {
    ticket_id: "missing",
    snapshot: { key: "CAPTURED-7", url: "https://example.test/old" },
  };
  assert.equal(resolveTicket(session, ref), undefined);
  assert.equal(
    resolveDelivery(session, {
      delivery_id: "missing",
      snapshot: {
        kind: "pull_request",
        status: "open",
        label: "Old",
        url: "https://example.test/old",
      },
    }),
    undefined,
  );
  const html = renderToStaticMarkup(
    <TimelineTicketLinks session={session} refs={[ref]} lang="es" />,
  );
  assert.doesNotMatch(html, /href=/);
  assert.match(html, /Publicado/);
});

test("two work-local references with equal published URLs stay distinct after one catalog update", () => {
  const { session } = completeFixture();
  const first = session.tickets[0],
    second = session.tickets[1];
  const refs = [first, second].map((t) => ({
    ticket_id: t.ticket_id,
    snapshot: { key: t.key, url: "https://example.test/old" },
  }));
  first.key = "FIRST-CORRECTED";
  first.url = "https://example.test/first";
  const html = renderToStaticMarkup(
    <TimelineTicketLinks session={session} refs={refs} lang="en" />,
  );
  assert.match(html, /FIRST-CORRECTED/);
  assert.ok(html.includes(second.key));
  assert.equal((html.match(/data-ticket-id=/g) ?? []).length, 2);
  assert.equal(resolveTicket(session, refs[1]), second);
});
