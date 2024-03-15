# Symphony agents

_(last edit: 3/14/2024)_

Symphony doesn’t mandate an agent to be installed on the targets it manages. Symphony prefers remote management interfaces when possible. On the other hand, if a target doesn’t support a remote management interface, a Symphony agent can be installed on the target so that Symphony control plane can manage it.

Symphony has four types of agents –Symphony target agent, Symphony proxy agent, Symphony poll agent, and a lightweight Symphony poll agent (code name Piccolo), as summarized in the following table:

| Agent Type | Protocol |
|--------|--------|
| Lightweight poll agent | HTTPS (outbound) |
| Poll agent | HTTPS (outbound) |
| Proxy agent | MQTT or HTTPS (inbound) |
| Target agent | HTTPS (outbound) |

> **NOTE:** Except for Piccolo, you can also combine agents into a combined agent, such has a Poll agent + Target agent.

## Related topics

* [Symphony lightweight polling agent (Piccolo)](./polling-agent.md)
* [Symphony polling agent](./polling-agent.md)
* [Symphony target agent](./target-agent.md)

