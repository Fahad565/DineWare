# DineWare Link

## 📋 Problem Statement
Modern safari lodges increasingly rely on cloud-based booking systems to manage guest logistics. However, remote camp kitchens operate in geographic zones plagued by erratic satellite links, dead cellular spots, and severe peak-hour network congestion, forcing Wi-Fi access to be restricted solely to guest lounges. Because back-of-house kitchen teams are physically cut off from internet connectivity, critical real-time updates—such as life-threatening guest food allergies, late cancellations, and dynamic multi-lodge itinerary changes—fail to propagate to chefs. This operational blind spot not only creates critical guest safety risks but also drives massive food wastage and financial loss, as remote kitchens over-prepare specialized ingredients and perishable supplies based on stale, un-updated manifests.

## 💎 Core Value Propositions
* **Zero-Data Reliability:** Kitchen-facing operational interactions run entirely over GSM signaling channels (USSD and SMS), requiring 0 Kbps of internet data bandwidth from the kitchen staff.
* **Granular Privacy & Security:** Sensitive health data (e.g., specific anaphylactic food allergies) is completely stripped from cleartext SMS notifications. Access is strictly controlled via authenticated USSD pins or tokenized secure web dashboards.
* **Reduced Food Wastage & Financial Protection:** Immediate offline awareness of real-time cancellations or itinerary shifts allows remote kitchens to dynamically throttle perishable food prep, optimizing supply chains that rely on infrequent bush deliveries.
* **Hybrid Connectivity Architecture:** Intelligently scales to provide high-fidelity rich dashboards if a camp has localized connectivity, while offering complete fallback functionality via raw telecom infrastructure when internet links fail entirely.

## 🛠️ Solution Methodology
DineWare bridges the operational gap between connected safari tour operators and disconnected bush kitchens through a two-tiered notification and retrieval architecture:
1. **Asynchronous Date-Staggered SMS Triggers:** A backend engine constantly monitors multi-lodge guest itineraries and fires a targeted, privacy-compliant Africa's Talking bulk SMS alert to a lodge's head chef exactly **72 hours prior to arrival**. This SMS contains *only* non-sensitive tracking markers: property-specific itinerary reference, guest count, and arrival date. No data links or web URLs are required.
2. **Secure Offline GSM Manifest Queries:** Upon receiving the SMS trigger, the chef dials an authenticated shortcode (`*384*44#`) using the standard GSM signaling channel. The request maps through Africa's Talking USSD gateway to our highly concurrent Go-engine backend, validating the kitchen's identity and pulling fresh guest manifests directly from Neon PostgreSQL, keeping sensitive health profiles secure without an active internet setup.

## 🧰 Technology Stack
* **Frontend:** Next.js (TypeScript) — Managed web platform for Tour Operators to register and manage guest vectors.
* **Backend:** Golang (v1.27) — High-concurrency engine handling Africa's Talking integrations and database state machine routines.
* **Database:** Neon PostgreSQL — Serverless relational layer hosting live operational manifests and audit records.
* **Communications Gateway:** Africa's Talking APIs (SMS & USSD pipelines).
* **Environments:** GitHub Codespaces (Core Dev Environment) + Google Jules (Project Spin Up).
* **Deployment System:** Render via `render.yaml` blueprints executing multi-stage Docker builds natively via `dockerContext`.

---

