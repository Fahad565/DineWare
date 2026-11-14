
## 🏃🏃 Development Roadmap (Sprint Breakdown)
All feature implementation progress is logged sequentially under the `docs/sprints/` subdirectory.

### 🏁 Sprint 1: Local Foundations & Environment Core (Days 1–3)
* **Goal:** Configure developer environment baselines and map Docker topologies.
* **Deliverables:**
  * Initialize monorepo structure in GitHub Codespaces.
  * Construct Next.js frontend scaffold and Golang v1.22 backend skeleton.
  * Formulate `render.yaml` Infrastructure-as-Code blueprint file.
  * Define localized `Dockerfile` configurations for both packages.
  * Establish connection pools to Neon PostgreSQL database.

### 🏁 Sprint 2: Core Engine Data Models & Web Portal (Days 4–7)
* **Goal:** Model relational dependencies and construct the operator ingestion interface.
* **Deliverables:**
  * Deploy migration schemas for `itineraries`, `itinerary_stops`, and `guests`.
  * Establish JWT-secured endpoints in the Go API layer.
  * Build the Next.js Tour Operator form to intake itinerary legs, guest details, and severe allergies.
  * Validate end-to-end data mutations from the web portal into the database layer.

### 🏁 Sprint 3: Privacy-Compliant SMS Engine & Staggered Dispatcher (Days 8–11)
* **Goal:** Code the backend asynchronous routing loops and connect Africa's Talking SMS API.
* **Deliverables:**
  * Implement the time-delta worker routines in Go executing cron-like scanning of check-in arrays.
  * Integrate Africa's Talking Outbound SMS client.
  * Enforce strict compliance filters ensuring SMS payloads *only* output the non-sensitive parameters (reference, guest count, arrival date, prompt) with **no cleartext health metrics**.
  * Validate that multi-lodge itineraries cleanly dispatch distinct staggered triggers sequentially.

### 🏁 Sprint 4: Authenticated USSD Manifest Pipeline (Days 12–15)
* **Goal:** Construct the zero-data offline lookup mechanics via GSM.
* **Deliverables:**
  * Open HTTP POST webhook endpoints for the Africa's Talking USSD protocol.
  * Implement text-string state routing (`CON` / `END`) within the Go application core.
  * Build security checks verifying incoming phone records against registered camp keys.
  * Render secure screens inside the USSD session showing detailed allergy specs and manifest properties.
  * Implement the explicit USSD confirmation loop changing state variables to true in PostgreSQL.

### 🏁 Sprint 5: System Integration, Failover Tests & Render Launch (Days 16–20)
* **Goal:** Perform absolute environment execution audits and execute deployment scripts.
* **Deliverables:**
  * Simulate high-congestion web data overrides during mid-itinerary paths to verify downstream data consistency.
  * Verify full USSD pipeline query speeds remain under the 2.5-second telco gateway ceiling.
  * Deploy the monorepo platform live to Render via the declarative blueprint system.
  * Conduct a live end-to-end sandbox demonstration showing Next.js data mutations triggering multi-channel GSM delivery.