---
title: Privacy Policy
description: Privacy Policy for SaveToInk
layout: ../layouts/MarkdownLayout.astro
---

# Privacy Policy

**Effective date:** March 9, 2026
**Last updated:** March 9, 2026

SaveToInk ("we", "us", or "our") is an open-source tool that converts web articles to EPUB and delivers them to e-readers. This policy explains what data we collect, why, and how we handle it.

---

## 1. What we collect

### Web App (app.saveto.ink) and Browser Extensions

- **URLs you explicitly submit.** When you click the SaveToInk button, the URL of the current tab is sent to the SaveToInk API for processing. No other browsing data is collected.
- **Authentication token.** Stored locally in browser storage to keep you signed in. Never shared with third parties.
- **Account information.** Your email address, collected via Auth0 at sign-up.
- **Delivery addresses.** E-reader email addresses (e.g. your Kindle address) you configure in account settings, used solely to deliver EPUBs.
- **Send history.** A log of articles delivered to your e-reader, used to calculate sends quota, display your history and for debugging delivery issues.

### CLI

- **Nothing beyond what you configure.** The CLI reads a local config file and communicates only with the SaveToInk API (or your self-hosted instance). No telemetry is collected.

### Self-hosted instances

If you run SaveToInk on your own infrastructure, all data stays on your servers. We have no access to it.

---

## 2. What we do not collect

- Browsing history
- Page content from pages you have not explicitly submitted
- Device identifiers or fingerprints
- Location data
- Any data from minors (our service is not directed at children under 13)

---

## 3. How we use your data

We use collected data only to:

- Convert submitted URLs to EPUBs and deliver them to your configured e-reader
- Maintain your article library and send history
- Authenticate your account
- Diagnose errors and improve reliability

We do not sell your data. We do not use it for advertising.

---

## 4. Third-party services

| Service             | Purpose                           | Their privacy policy                                                  |
| ------------------- | --------------------------------- | --------------------------------------------------------------------- |
| Amazon Web Services | Infrastructure (Lambda, DynamoDB) | [aws.amazon.com/privacy](https://aws.amazon.com/privacy)              |
| Auth0               | Authentication                    | [auth0.com/privacy](https://auth0.com/privacy)                        |
| Cloudflare          | Infrastructure (Workers)          | [cloudflare.com/privacy](https://www.cloudflare.com/privacy/)         |
| Mailjet             | E-reader email delivery           | [mailjet.com/privacy-policy](https://www.mailjet.com/privacy-policy/) |
| Sentry              | Error monitoring and telemetry    | [sentry.io/privacy](https://sentry.io/privacy)                        |

These services receive only the data necessary to perform their function.

---

## 5. Data retention

- **Article content and send history** — retained until you delete them or close your account.
- **Account data** — deleted within 30 days of account deletion upon request.
- **Error logs** — retained for 90 days.

---

## 6. Your rights

You can at any time:

- Delete individual articles or your entire library
- Delete your account by contacting us at **privacy@saveto.ink**

If you are in the EU or UK, you have additional rights under GDPR/UK GDPR, including the right to access, rectify, erase, or port your data, and to lodge a complaint with your local supervisory authority.

---

## 7. Security

All data in transit is encrypted via TLS. Data at rest is stored in AWS infrastructure with encryption enabled. We follow reasonable security practices, but no system is perfectly secure. Each and every component of the application is open source and auditable at [the GitHub repository](https://github.com/savetoink/savetoink)

---

## 8. Changes to this policy

We may update this policy as the product evolves. We will update the "Last updated" date at the top. Continued use of SaveToInk after changes constitutes acceptance of the updated policy.

---

## 9. Contact

Questions or requests: **privacy@saveto.ink**
