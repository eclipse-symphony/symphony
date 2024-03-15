# Symphony polling agent

_(last edit: 3/14/2024)_

Symphony polling agent connects to the Symphony control plane through a single outbound HTTPS connection. It reports target current states and retrieves the new desired states from the control plane. Then, it runs a local reconciliation process. The polling agent is used in conjunction with Staging Target provider, which stages the desired state on the control plane itself instead of pushing it out to the target. The polling agent periodically polls the control plane for updated desired states.


## Related topics

* [Symphony lightweight polling agent (Piccolo)](./piccolo-agent.md)
