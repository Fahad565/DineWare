
Product Requirements Document (PRD): DineWare


1. Introduction & Objectives

DineWare is an offline-first dietary and itinerary synchronization engine built to bridge the digital gap between safari tour operators and remote camp kitchens.
• Primary Objective: Guarantee 100% delivery of critical guest dietary requirements, allergy manifests, and group headcount updates to remote lodge kitchens without relying on internet data infrastructure.
• Secondary Objective: Eliminate avoidable food wastage and financial loss caused by over-preparing perishable ingredients against stale, un-updated manifests.
• Core Strategy: Pair staggered, privacy-safe SMS notifications with an authenticated USSD lookup channel via Africa’s Talking APIs, alongside a secure web portal for properties that have active connectivity.

2. User Personas

• The Tour Operator (Web User): Uses the Next.js portal to create multi-stop safari itineraries, register group headcounts, log dietary/allergy restrictions, and update trip details in real time.
• The Camp Head Chef / Kitchen Manager (USSD & Web User): Operates back-of-house in areas with erratic or zero Wi-Fi. Uses a basic GSM phone to dial USSD for offline manifest retrieval, or uses the secure web portal when lounge/admin internet is available.

3. Functional Requirements

• Staggered SMS Dispatch Engine (Privacy-Safe):
	• The Golang backend engine evaluates active multi-lodge itineraries and dispatches an automated SMS alert to each lodge's designated chef 72 hours prior to guest arrival.
	• Strict Data Redaction Rule: Ordinary SMS messages must never contain sensitive health or dietary details (such as specific allergies). Each SMS contains only:
		1. Property-specific itinerary reference code.
		2. Number of arriving guests.
		3. Scheduled arrival date.
		4. A prompt to retrieve the full manifest via USSD (*384*44#) or the secure web portal.
• Dual-Channel Manifest Retrieval (Offline USSD + Online Web):
	• Offline Channel (USSD): Chefs dial the dedicated shortcode (*384*44#), authenticate their session, and enter the itinerary reference code to view detailed guest allergies, dietary needs, and meal counts without mobile data.
	• Online Channel (Secure Web Page): For camps with intermittent or lounge connectivity, an authenticated web view provides a richer kitchen manifest dashboard.
• Dynamic Multi-Lodge Propagation:
	• When a tour operator modifies guest counts, cancellations, or allergy records mid-safari, downstream lodges whose 72-hour dispatch window has not yet triggered automatically receive the updated manifest.
	• If a lodge has already received its initial alert, high-priority changes trigger an urgent update SMS prompting the chef to re-check the manifest.
• Chef Acknowledgment & Confirmation Loop:
	• The USSD flow and web portal require the chef to confirm receipt of the manifest, updating the itinerary status to "Confirmed" on the tour operator's dashboard for auditability.

4. Non-Functional Requirements

• Zero-Data Reliability: Core kitchen alerting and manifest lookup must run 100% over GSM signaling channels (SMS and USSD), requiring 0 Kbps of internet bandwidth.
• Privacy & Health Data Protection: Sensitive dietary/allergy records are treated as protected health-related information and are strictly gated behind authenticated USSD sessions or secured HTTPS web endpoints.
• Low USSD Latency: USSD menu responses must compile and return within 2.5 seconds to prevent telco gateway session timeouts.
• Containerized Portability: Both the Next.js frontend and Golang 1.27 backend must run as reproducible Docker containers backed by Neon PostgreSQL and deployed via Render (render.yaml).
