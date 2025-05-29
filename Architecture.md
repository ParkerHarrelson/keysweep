                                                     +-----------------------+
                                                     |   Slack / Teams       |
                                                     |  (Webhook JSON)       |
                                                     +-----------▲-----------+
                                                                 |
                                         4. Slack notify         |
                                                                 |
+------------------+          HTTPS/JSON            +------------+-------------+
| GitHub Action    |--------------------------------> KeySweep Control‑Plane   |
| 1. scanner step  |                                 | (Spring Boot 3 – Kotlin)|
|------------------|                                 |------------------------ |
| • wrapper.sh (bash)                               | • Policy Engine (OPA)  |
| • keysweep‑scanner (Go) ← gitleaks binary (Go)     | • Rotator Pool        |
|   – diff → JSON findings                           |   ├─ aws‑rotator.kt   |
|   – signs payload (ed25519)                        |   └─ github‑rotator.kt|
+---------+--------+                                 | • PR Builder (JGit ‑ Java)|
          |                                          | • Audit Log (JPA)        |
          | GraphQL / REST                           +------+----+------^------+
2. PR create / 5. merge fix PR                              |           |
          |                                                 |           |
          |                              3. Rotate (SDK)    |           |
          |                                                 |           |
+---------v--------+                                 +------+-----------+------+
| GitHub API       |                                 |  Secret Stores          |
|  (Octokit REST)  |                                 |  ├─ AWS SecretsMgr SDK |
|                  |                                 |  └─ GitHub Tokens API  |
+------------------+                                 +------------------------+
